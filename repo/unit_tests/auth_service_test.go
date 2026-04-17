package unit_tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/localinsights/portal/internal/config"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/errs"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/captcha"
	"github.com/localinsights/portal/internal/pkg/hash"
	"github.com/localinsights/portal/internal/pkg/jwt"
	"github.com/localinsights/portal/internal/repository"
	"github.com/localinsights/portal/internal/service"
)

// --- In-memory fakes for the four repositories AuthService depends on ----

type fakeUserRepo struct {
	byUsername map[string]*model.User
	byEmail    map[string]*model.User
	byID       map[uint64]*model.User
	nextID     uint64
	createErr  error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byUsername: map[string]*model.User{},
		byEmail:    map[string]*model.User{},
		byID:       map[uint64]*model.User{},
		nextID:     1,
	}
}

func (r *fakeUserRepo) Create(ctx context.Context, u *model.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	u.ID = r.nextID
	r.nextID++
	r.byUsername[u.Username] = u
	r.byEmail[u.Email] = u
	r.byID[u.ID] = u
	return nil
}
func (r *fakeUserRepo) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	if u, ok := r.byID[id]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}
func (r *fakeUserRepo) GetByUUID(ctx context.Context, uuid string) (*model.User, error) {
	for _, u := range r.byID {
		if u.UUID == uuid {
			return u, nil
		}
	}
	return nil, nil
}
func (r *fakeUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return r.byUsername[username], nil
}
func (r *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.byEmail[email], nil
}
func (r *fakeUserRepo) Update(ctx context.Context, u *model.User) error { return nil }
func (r *fakeUserRepo) List(ctx context.Context, page repository.Pagination) ([]*model.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeUserRepo) UpdateRole(ctx context.Context, id uint64, role model.Role) error { return nil }
func (r *fakeUserRepo) SetActive(ctx context.Context, id uint64, active bool) error      { return nil }

type fakePrefsRepo struct {
	upserted []*model.UserPreferences
	err      error
}

func (r *fakePrefsRepo) Upsert(ctx context.Context, p *model.UserPreferences) error {
	if r.err != nil {
		return r.err
	}
	r.upserted = append(r.upserted, p)
	return nil
}
func (r *fakePrefsRepo) GetByUserID(ctx context.Context, userID uint64) (*model.UserPreferences, error) {
	return nil, nil
}

type fakeLoginAttemptRepo struct {
	created   []*model.LoginAttempt
	failedBy  map[string]int
}

func newFakeLoginAttemptRepo() *fakeLoginAttemptRepo {
	return &fakeLoginAttemptRepo{failedBy: map[string]int{}}
}
func (r *fakeLoginAttemptRepo) Create(ctx context.Context, a *model.LoginAttempt) error {
	r.created = append(r.created, a)
	return nil
}
func (r *fakeLoginAttemptRepo) CountRecentFailed(ctx context.Context, email, ip string, window time.Duration) (int, error) {
	return r.failedBy[email], nil
}

type fakeRefreshTokenRepo struct {
	tokens  map[string]uint64
	revoked map[string]bool
}

func newFakeRefreshTokenRepo() *fakeRefreshTokenRepo {
	return &fakeRefreshTokenRepo{tokens: map[string]uint64{}, revoked: map[string]bool{}}
}
func (r *fakeRefreshTokenRepo) Create(ctx context.Context, userID uint64, tokenHash string, expiresAt time.Time) error {
	r.tokens[tokenHash] = userID
	return nil
}
func (r *fakeRefreshTokenRepo) GetByHash(ctx context.Context, tokenHash string) (uint64, error) {
	if r.revoked[tokenHash] {
		return 0, errors.New("revoked")
	}
	if id, ok := r.tokens[tokenHash]; ok {
		return id, nil
	}
	return 0, errors.New("not found")
}
func (r *fakeRefreshTokenRepo) Revoke(ctx context.Context, tokenHash string) error {
	r.revoked[tokenHash] = true
	delete(r.tokens, tokenHash)
	return nil
}
func (r *fakeRefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID uint64) error {
	return nil
}

// --- Helpers --------------------------------------------------------------

func newAuthService(t *testing.T, users *fakeUserRepo, attempts *fakeLoginAttemptRepo) (*service.AuthService, *fakeRefreshTokenRepo, *captcha.Store) {
	t.Helper()
	jwtMgr := jwt.NewManager(config.JWTConfig{
		Secret:     "test-secret-must-be-long-enough-for-hs256-signing",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	})
	prefs := &fakePrefsRepo{}
	tokens := newFakeRefreshTokenRepo()
	capStore := captcha.NewStore()
	svc := service.NewAuthService(users, prefs, attempts, tokens, jwtMgr, 5, 15*time.Minute, capStore)
	return svc, tokens, capStore
}

func mustHash(t *testing.T, pw string) string {
	t.Helper()
	h, err := hash.HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	return h
}

// --- Register -------------------------------------------------------------

func TestAuthServiceRegisterCreatesUserAndDefaultPrefs(t *testing.T) {
	users := newFakeUserRepo()
	attempts := newFakeLoginAttemptRepo()
	svc, _, _ := newAuthService(t, users, attempts)

	got, err := svc.Register(context.Background(), &request.RegisterRequest{
		Username: "alice", Email: "a@example.com", Password: "SecurePass1",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if got == nil || got.Username != "alice" || got.Email != "a@example.com" {
		t.Fatalf("unexpected user: %+v", got)
	}
	if got.Role != model.RoleRegularUser {
		t.Errorf("default role = %q, want regular_user", got.Role)
	}
	if !got.IsActive {
		t.Error("new user should be active by default")
	}
	if got.PasswordHash == "" || got.PasswordHash == "SecurePass1" {
		t.Error("password should be hashed, not stored plaintext")
	}
}

func TestAuthServiceRegisterDuplicateUsernameFails(t *testing.T) {
	users := newFakeUserRepo()
	users.byUsername["taken"] = &model.User{Username: "taken"}
	svc, _, _ := newAuthService(t, users, newFakeLoginAttemptRepo())

	_, err := svc.Register(context.Background(), &request.RegisterRequest{
		Username: "taken", Email: "new@example.com", Password: "SecurePass1",
	})
	if !errs.Is(err, errs.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestAuthServiceRegisterDuplicateEmailFails(t *testing.T) {
	users := newFakeUserRepo()
	users.byEmail["a@example.com"] = &model.User{Email: "a@example.com"}
	svc, _, _ := newAuthService(t, users, newFakeLoginAttemptRepo())

	_, err := svc.Register(context.Background(), &request.RegisterRequest{
		Username: "newuser", Email: "a@example.com", Password: "SecurePass1",
	})
	if !errs.Is(err, errs.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

// --- Login ----------------------------------------------------------------

func TestAuthServiceLoginSuccess(t *testing.T) {
	users := newFakeUserRepo()
	u := &model.User{
		UUID:         "u-uuid-1",
		Username:     "bob",
		Email:        "b@example.com",
		PasswordHash: mustHash(t, "SecurePass1"),
		Role:         model.RoleRegularUser,
		IsActive:     true,
	}
	_ = users.Create(context.Background(), u)
	svc, tokens, _ := newAuthService(t, users, newFakeLoginAttemptRepo())

	access, refresh, gotUser, err := svc.Login(context.Background(), &request.LoginRequest{
		Username: "bob", Password: "SecurePass1",
	}, "127.0.0.1")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Error("tokens should be non-empty on success")
	}
	if gotUser == nil || gotUser.Username != "bob" {
		t.Errorf("unexpected user: %+v", gotUser)
	}
	if len(tokens.tokens) != 1 {
		t.Errorf("expected 1 refresh token stored, got %d", len(tokens.tokens))
	}
}

func TestAuthServiceLoginWrongPasswordFails(t *testing.T) {
	users := newFakeUserRepo()
	u := &model.User{
		Username:     "bob",
		Email:        "b@example.com",
		PasswordHash: mustHash(t, "correct-password"),
		Role:         model.RoleRegularUser,
		IsActive:     true,
	}
	_ = users.Create(context.Background(), u)
	attempts := newFakeLoginAttemptRepo()
	svc, _, _ := newAuthService(t, users, attempts)

	_, _, _, err := svc.Login(context.Background(), &request.LoginRequest{
		Username: "bob", Password: "wrong-password",
	}, "127.0.0.1")
	if !errs.Is(err, errs.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	// Login attempt must be recorded as a failure so the captcha threshold
	// can track it.
	if len(attempts.created) != 1 || attempts.created[0].Success {
		t.Errorf("expected 1 failed login attempt recorded, got %+v", attempts.created)
	}
}

func TestAuthServiceLoginUnknownUserFails(t *testing.T) {
	svc, _, _ := newAuthService(t, newFakeUserRepo(), newFakeLoginAttemptRepo())
	_, _, _, err := svc.Login(context.Background(), &request.LoginRequest{
		Username: "does-not-exist", Password: "whatever",
	}, "1.2.3.4")
	if !errs.Is(err, errs.ErrUnauthorized) {
		t.Errorf("expected unauthorized for unknown user, got %v", err)
	}
}

func TestAuthServiceLoginInactiveAccountForbidden(t *testing.T) {
	users := newFakeUserRepo()
	u := &model.User{
		Username:     "zombie",
		PasswordHash: mustHash(t, "SecurePass1"),
		Role:         model.RoleRegularUser,
		IsActive:     false,
	}
	_ = users.Create(context.Background(), u)
	svc, _, _ := newAuthService(t, users, newFakeLoginAttemptRepo())

	_, _, _, err := svc.Login(context.Background(), &request.LoginRequest{
		Username: "zombie", Password: "SecurePass1",
	}, "1.2.3.4")
	if !errs.Is(err, errs.ErrForbidden) {
		t.Errorf("expected forbidden for inactive account, got %v", err)
	}
}

func TestAuthServiceLoginCaptchaRequiredAfterThreshold(t *testing.T) {
	users := newFakeUserRepo()
	u := &model.User{
		Username:     "charlie",
		PasswordHash: mustHash(t, "SecurePass1"),
		Role:         model.RoleRegularUser,
		IsActive:     true,
	}
	_ = users.Create(context.Background(), u)
	attempts := newFakeLoginAttemptRepo()
	attempts.failedBy["charlie"] = 5 // at the threshold
	svc, _, _ := newAuthService(t, users, attempts)

	// No captcha → ErrCaptchaRequired.
	_, _, _, err := svc.Login(context.Background(), &request.LoginRequest{
		Username: "charlie", Password: "SecurePass1",
	}, "1.2.3.4")
	if !errs.Is(err, errs.ErrCaptchaRequired) {
		t.Errorf("expected captcha required, got %v", err)
	}
}

func TestAuthServiceLoginCaptchaInvalid(t *testing.T) {
	users := newFakeUserRepo()
	u := &model.User{
		Username:     "diana",
		PasswordHash: mustHash(t, "SecurePass1"),
		Role:         model.RoleRegularUser,
		IsActive:     true,
	}
	_ = users.Create(context.Background(), u)
	attempts := newFakeLoginAttemptRepo()
	attempts.failedBy["diana"] = 10
	svc, _, _ := newAuthService(t, users, attempts)

	_, _, _, err := svc.Login(context.Background(), &request.LoginRequest{
		Username: "diana", Password: "SecurePass1",
		CaptchaID: "bogus-id", CaptchaToken: "bogus-answer",
	}, "1.2.3.4")
	if !errs.Is(err, errs.ErrCaptchaInvalid) {
		t.Errorf("expected captcha invalid, got %v", err)
	}
}

// --- RefreshAccessToken ---------------------------------------------------

func TestAuthServiceRefreshIssuesNewAccessTokenWhenRefreshTokenIsValid(t *testing.T) {
	users := newFakeUserRepo()
	u := &model.User{
		Username:     "refresh-user",
		PasswordHash: mustHash(t, "SecurePass1"),
		Role:         model.RoleRegularUser,
		IsActive:     true,
	}
	_ = users.Create(context.Background(), u)
	svc, _, _ := newAuthService(t, users, newFakeLoginAttemptRepo())

	_, refresh, _, err := svc.Login(context.Background(), &request.LoginRequest{
		Username: "refresh-user", Password: "SecurePass1",
	}, "1.2.3.4")
	if err != nil {
		t.Fatalf("seed login failed: %v", err)
	}

	newAccess, gotUser, err := svc.RefreshAccessToken(context.Background(), refresh)
	if err != nil {
		t.Fatalf("RefreshAccessToken returned error: %v", err)
	}
	if newAccess == "" || gotUser == nil || gotUser.Username != "refresh-user" {
		t.Errorf("unexpected refresh result: access=%q user=%+v", newAccess, gotUser)
	}
}

func TestAuthServiceRefreshWithInvalidTokenFails(t *testing.T) {
	svc, _, _ := newAuthService(t, newFakeUserRepo(), newFakeLoginAttemptRepo())
	_, _, err := svc.RefreshAccessToken(context.Background(), "definitely.not.a.jwt")
	if !errs.Is(err, errs.ErrUnauthorized) {
		t.Errorf("expected unauthorized, got %v", err)
	}
}

// --- Logout ---------------------------------------------------------------

func TestAuthServiceLogoutRevokesRefreshToken(t *testing.T) {
	users := newFakeUserRepo()
	u := &model.User{
		Username:     "logout-user",
		PasswordHash: mustHash(t, "SecurePass1"),
		Role:         model.RoleRegularUser,
		IsActive:     true,
	}
	_ = users.Create(context.Background(), u)
	svc, tokens, _ := newAuthService(t, users, newFakeLoginAttemptRepo())

	_, refresh, _, _ := svc.Login(context.Background(), &request.LoginRequest{
		Username: "logout-user", Password: "SecurePass1",
	}, "1.2.3.4")

	if len(tokens.tokens) != 1 {
		t.Fatalf("expected 1 token before logout, got %d", len(tokens.tokens))
	}

	if err := svc.Logout(context.Background(), refresh); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}
	if len(tokens.tokens) != 0 {
		t.Errorf("token should be removed after logout; got %d tokens left", len(tokens.tokens))
	}
	// Refresh with the logged-out token must now fail.
	_, _, err := svc.RefreshAccessToken(context.Background(), refresh)
	if !errs.Is(err, errs.ErrUnauthorized) {
		t.Errorf("refreshing a logged-out token should fail; got %v", err)
	}
}

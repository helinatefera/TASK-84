export enum Role {
  User = 'regular_user',
  Moderator = 'moderator',
  Analyst = 'product_analyst',
  Admin = 'admin',
}

export enum ReportCategory {
  Spam = 'spam',
  Harassment = 'harassment',
  Misinformation = 'misinformation',
  Inappropriate = 'inappropriate',
  Copyright = 'copyright',
  Other = 'other',
}

export enum ReportStatus {
  Pending = 'pending',
  InReview = 'in_review',
  Resolved = 'resolved',
  Dismissed = 'dismissed',
}

export enum FraudStatus {
  Normal = 'normal',
  Suspected = 'suspected_fraud',
  Confirmed = 'confirmed_fraud',
  Cleared = 'cleared',
}

export enum ExperimentStatus {
  Draft = 'draft',
  Running = 'running',
  Paused = 'paused',
  Completed = 'completed',
  RolledBack = 'rolled_back',
}

export enum ConfidenceState {
  InsufficientData = 'insufficient_data',
  Monitoring = 'monitoring',
  RecommendKeep = 'recommend_keep',
  RecommendRollback = 'recommend_rollback',
}

export enum Sentiment {
  Positive = 'positive',
  Neutral = 'neutral',
  Negative = 'negative',
}

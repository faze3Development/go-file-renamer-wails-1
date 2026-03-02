import React from 'react'

interface Props {
  title?: string
  message?: string
}

export default function LoadingBoundary({
  title = 'Loading Application',
  message = 'Initializing components...',
}: Props) {
  return (
    <div className="loading-boundary">
      <div className="loading-skeleton-layout">
        {/* Fake sidebar */}
        <div className="skeleton-sidebar">
          <div className="skeleton skeleton-logo" />
          <div className="skeleton skeleton-badge" />
          <div style={{ padding: '12px 8px', display: 'flex', flexDirection: 'column', gap: '6px' }}>
            {[1, 2, 3, 4, 5].map(i => (
              <div key={i} className="skeleton skeleton-nav-item" />
            ))}
          </div>
        </div>

        {/* Fake main content */}
        <div className="skeleton-main">
          <div className="skeleton-card-group">
            <div className="skeleton skeleton-card-header" />
            <div className="skeleton skeleton-card-body" />
          </div>
          <div className="skeleton-card-grid">
            <div className="skeleton-card-group">
              <div className="skeleton skeleton-card-header" />
              <div className="skeleton skeleton-card-body short" />
            </div>
            <div className="skeleton-card-group">
              <div className="skeleton skeleton-card-header" />
              <div className="skeleton skeleton-card-body short" />
            </div>
          </div>
          <div className="skeleton-loading-label">
            <div className="loading-spinner-sm" />
            <span>{title} — {message}</span>
          </div>
        </div>
      </div>
    </div>
  )
}

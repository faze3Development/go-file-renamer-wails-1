import React, { Component } from 'react'
import { CircleAlert, RefreshCw, X } from 'lucide-react'

interface Props {
  errorMessage?: string
  onRetry?: () => void
  onDismiss?: () => void
}

interface State {
  hasError: boolean
  error?: Error
}

export default class ErrorBoundary extends Component<React.PropsWithChildren<Props>, State> {
  constructor(props: React.PropsWithChildren<Props>) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('ErrorBoundary caught:', error, info)
  }

  render() {
    const { errorMessage, onRetry, onDismiss, children } = this.props
    const { hasError, error } = this.state
    const displayMessage = errorMessage || (hasError ? error?.message || 'An unexpected error occurred.' : '')

    if (!displayMessage && !hasError) return children

    return (
      <div className="error-boundary">
        <div className="error-content">
          <div className="error-icon">
            <CircleAlert size={48} />
          </div>
          <h2>Something went wrong</h2>
          <p className="error-message">{displayMessage}</p>
          <div className="error-actions">
            <button
              className="retry-btn"
              onClick={() => {
                this.setState({ hasError: false, error: undefined })
                onRetry?.()
              }}
            >
              <RefreshCw size={16} />
              Retry
            </button>
            <button
              className="secondary-btn"
              onClick={() => {
                this.setState({ hasError: false, error: undefined })
                onDismiss?.()
              }}
            >
              <X size={16} />
              Dismiss
            </button>
          </div>
        </div>
      </div>
    )
  }
}

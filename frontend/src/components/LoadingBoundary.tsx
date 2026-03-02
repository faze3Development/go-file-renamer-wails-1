import React from 'react'

interface Props {
  title?: string
  message?: string
}

export default function LoadingBoundary({
  title = 'Loading Application',
  message = 'Initializing components and loading data...',
}: Props) {
  return (
    <div className="loading-boundary">
      <div className="loading-content">
        <div className="loading-spinner" />
        <h2>{title}</h2>
        <p>{message}</p>
      </div>
    </div>
  )
}

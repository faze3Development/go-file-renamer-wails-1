import React from 'react'

interface CardProps {
  compact?: boolean
  fullWidth?: boolean
  header?: React.ReactNode
  children: React.ReactNode
}

export default function Card({ compact = false, fullWidth = false, header, children }: CardProps) {
  const classes = ['card', compact ? 'compact' : '', fullWidth ? 'full-width' : ''].filter(Boolean).join(' ')
  return (
    <div className={classes}>
      {header}
      {children}
    </div>
  )
}

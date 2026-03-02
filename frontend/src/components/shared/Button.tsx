import React from 'react'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  ariaLabel?: string
}

export default function Button({ variant = 'primary', ariaLabel, className = '', children, ...props }: ButtonProps) {
  return (
    <button
      className={`btn ${variant}${className ? ` ${className}` : ''}`}
      aria-label={ariaLabel || undefined}
      {...props}
    >
      {children}
    </button>
  )
}

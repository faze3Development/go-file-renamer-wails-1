import React from 'react'

interface ToggleProps {
  checked: boolean
  disabled?: boolean
  id?: string
  ariaLabel?: string
  size?: 'sm' | 'md' | 'lg'
  onChange?: (checked: boolean) => void
}

export default function Toggle({ checked, disabled = false, id, ariaLabel, size = 'md', onChange }: ToggleProps) {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange?.(e.target.checked)
  }

  return (
    <label className="toggle" aria-label={ariaLabel} aria-disabled={disabled}>
      <input
        type="checkbox"
        role="switch"
        id={id}
        disabled={disabled}
        checked={checked}
        onChange={handleChange}
      />
      <span className="slider" data-size={size} aria-hidden="true" />
    </label>
  )
}

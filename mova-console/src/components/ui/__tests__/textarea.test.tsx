import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Textarea } from '../textarea'

describe('Textarea', () => {
  it('renders correctly', () => {
    render(<Textarea placeholder="Enter text..." />)
    
    const textarea = screen.getByPlaceholderText('Enter text...')
    expect(textarea).toBeInTheDocument()
    expect(textarea.tagName).toBe('TEXTAREA')
  })

  it('applies default classes', () => {
    render(<Textarea data-testid="textarea" />)
    
    const textarea = screen.getByTestId('textarea')
    expect(textarea).toHaveClass(
      'flex',
      'min-h-[80px]',
      'w-full',
      'rounded-md',
      'border',
      'border-input',
      'bg-background',
      'px-3',
      'py-2',
      'text-sm'
    )
  })

  it('handles value changes', async () => {
    const handleChange = jest.fn()
    const user = userEvent.setup()
    
    render(<Textarea onChange={handleChange} />)
    
    const textarea = screen.getByRole('textbox')
    await user.type(textarea, 'Hello World')
    
    expect(handleChange).toHaveBeenCalled()
    expect(textarea).toHaveValue('Hello World')
  })

  it('can be disabled', () => {
    render(<Textarea disabled data-testid="textarea" />)
    
    const textarea = screen.getByTestId('textarea')
    expect(textarea).toBeDisabled()
    expect(textarea).toHaveClass('disabled:cursor-not-allowed', 'disabled:opacity-50')
  })

  it('applies custom className', () => {
    render(<Textarea className="custom-textarea" data-testid="textarea" />)
    
    const textarea = screen.getByTestId('textarea')
    expect(textarea).toHaveClass('custom-textarea')
  })

  it('forwards ref correctly', () => {
    const ref = jest.fn()
    render(<Textarea ref={ref} />)
    
    expect(ref).toHaveBeenCalled()
  })

  it('supports controlled input', async () => {
    const TestComponent = () => {
      const [value, setValue] = React.useState('')
      return (
        <Textarea 
          value={value} 
          onChange={(e) => setValue(e.target.value)}
          data-testid="controlled-textarea"
        />
      )
    }

    const user = userEvent.setup()
    render(<TestComponent />)
    
    const textarea = screen.getByTestId('controlled-textarea')
    await user.type(textarea, 'Controlled text')
    
    expect(textarea).toHaveValue('Controlled text')
  })

  it('supports different sizes via className', () => {
    render(<Textarea className="min-h-[200px]" data-testid="large-textarea" />)
    
    const textarea = screen.getByTestId('large-textarea')
    expect(textarea).toHaveClass('min-h-[200px]')
  })

  it('supports placeholder text', () => {
    render(<Textarea placeholder="Type your message here..." />)
    
    const textarea = screen.getByPlaceholderText('Type your message here...')
    expect(textarea).toBeInTheDocument()
  })
})

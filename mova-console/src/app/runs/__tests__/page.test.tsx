import { render, screen } from '@testing-library/react'
import RunsPage from '../page'

// Mock Link component
jest.mock('next/link', () => {
  return function MockLink({ children, href }: { children: React.ReactNode, href: string }) {
    return <a href={href}>{children}</a>
  }
})

describe('Runs Page', () => {
  it('renders without crashing', () => {
    render(<RunsPage />)
    
    // Should render some content
    expect(document.body).toBeInTheDocument()
  })

  it('shows loading state initially', () => {
    render(<RunsPage />)
    
    // Should have a loading spinner
    const spinners = screen.getAllByRole('generic').filter(el => 
      el.classList.contains('animate-spin')
    )
    expect(spinners.length).toBeGreaterThan(0)
  })

  it('handles component lifecycle', () => {
    const { unmount } = render(<RunsPage />)
    
    // Should unmount without errors
    expect(() => unmount()).not.toThrow()
  })
})
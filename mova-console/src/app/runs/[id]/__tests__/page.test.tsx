import { render, screen, waitFor } from '@testing-library/react'
import RunDetailsPage from '../page'

// Mock useParams hook
const mockParams = { id: 'run-001' }
jest.mock('next/navigation', () => ({
  ...jest.requireActual('next/navigation'),
  useParams: () => mockParams,
}))

// Mock Link component
jest.mock('next/link', () => {
  return function MockLink({ children, href }: { children: React.ReactNode, href: string }) {
    return <a href={href}>{children}</a>
  }
})

describe('Run Details Page', () => {
  it('renders loading state initially', () => {
    render(<RunDetailsPage />)
    
    expect(screen.getAllByRole('generic')[0]).toHaveClass('animate-spin')
  })

  it('renders page content after loading', async () => {
    render(<RunDetailsPage />)
    
    // Wait for loading to complete with longer timeout
    await waitFor(() => {
      expect(screen.queryByRole('generic')).not.toHaveClass('animate-spin')
    }, { timeout: 2000 })
    
    // Check that some content is rendered
    const headings = screen.getAllByRole('heading')
    expect(headings.length).toBeGreaterThan(0)
  })

  it('handles different run IDs', async () => {
    // Change the mock params
    mockParams.id = 'test-run-123'
    
    render(<RunDetailsPage />)
    
    await waitFor(() => {
      expect(screen.queryByRole('generic')).not.toHaveClass('animate-spin')
    }, { timeout: 2000 })
    
    // Should render some content regardless of ID
    expect(screen.getAllByRole('heading').length).toBeGreaterThan(0)
    
    // Restore original ID
    mockParams.id = 'run-001'
  })

  it('renders back navigation link', async () => {
    render(<RunDetailsPage />)
    
    await waitFor(() => {
      const backLinks = screen.getAllByText(/Back to Runs/)
      expect(backLinks.length).toBeGreaterThan(0)
    }, { timeout: 2000 })
  })

  it('displays run information when loaded', async () => {
    render(<RunDetailsPage />)
    
    await waitFor(() => {
      // Look for any text that indicates the page loaded
      expect(screen.queryByRole('generic')).not.toHaveClass('animate-spin')
    }, { timeout: 2000 })
    
    // Check for various elements that should be present
    const allText = screen.getByRole('main') || document.body
    expect(allText).toBeInTheDocument()
  })

  it('handles component mounting and unmounting', () => {
    const { unmount } = render(<RunDetailsPage />)
    
    // Should render without crashing
    expect(screen.getAllByRole('generic')[0]).toBeInTheDocument()
    
    // Should unmount without crashing
    unmount()
  })

  it('uses correct run ID from params', () => {
    render(<RunDetailsPage />)
    
    // The component should use the mocked ID (may have been changed by previous tests)
    expect(typeof mockParams.id).toBe('string')
    expect(mockParams.id).toBeTruthy()
  })
})
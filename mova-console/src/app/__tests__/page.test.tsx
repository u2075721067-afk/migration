import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Home from '../page'

// Mock the API module
const mockExecuteEnvelope = jest.fn()
jest.mock('@/lib/api', () => ({
  api: {
    executeEnvelope: mockExecuteEnvelope,
  },
}))

// Mock Link component  
jest.mock('next/link', () => {
  return function MockLink({ children, href }: { children: React.ReactNode, href: string }) {
    return <a href={href}>{children}</a>
  }
})

describe('Home Page', () => {
  beforeEach(() => {
    mockExecuteEnvelope.mockClear()
  })

  it('renders without crashing', () => {
    render(<Home />)
    
    expect(document.body).toBeInTheDocument()
  })

  it('renders the main heading', () => {
    render(<Home />)
    
    expect(screen.getByRole('heading', { name: 'MOVA Console' })).toBeInTheDocument()
  })

  it('renders the textarea for envelope input', () => {
    render(<Home />)
    
    const textarea = screen.getByRole('textbox')
    expect(textarea).toBeInTheDocument()
    expect(textarea).toHaveValue(expect.stringContaining('"intent": "demo-workflow"'))
  })

  it('renders the run button', () => {
    render(<Home />)
    
    expect(screen.getByRole('button', { name: /Run Envelope/ })).toBeInTheDocument()
  })

  it('allows editing the envelope', async () => {
    const user = userEvent.setup()
    render(<Home />)
    
    const textarea = screen.getByRole('textbox')
    
    await user.clear(textarea)
    await user.type(textarea, '{"test": "envelope"}')
    
    expect(textarea).toHaveValue('{"test": "envelope"}')
  })

  it('handles successful execution', async () => {
    const mockResult = { status: 'success', data: 'test result' }
    mockExecuteEnvelope.mockResolvedValue(mockResult)
    
    const user = userEvent.setup()
    render(<Home />)
    
    const runButton = screen.getByRole('button', { name: /Run Envelope/ })
    
    await user.click(runButton)
    
    await waitFor(() => {
      expect(mockExecuteEnvelope).toHaveBeenCalled()
    })
  })

  it('handles invalid JSON', async () => {
    const user = userEvent.setup()
    render(<Home />)
    
    const textarea = screen.getByRole('textbox')
    const runButton = screen.getByRole('button', { name: /Run Envelope/ })
    
    await user.clear(textarea)
    await user.type(textarea, '{"invalid": json}')
    
    await user.click(runButton)
    
    await waitFor(() => {
      expect(screen.getByText('"error": "Invalid JSON format"')).toBeInTheDocument()
    })
    
    expect(mockExecuteEnvelope).not.toHaveBeenCalled()
  })

  it('shows default result message initially', () => {
    render(<Home />)
    
    expect(screen.getByText('Click "Run Envelope" to see results here')).toBeInTheDocument()
  })
})
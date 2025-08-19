import { render, screen } from '@testing-library/react'
import { 
  Table, 
  TableBody, 
  TableCaption, 
  TableCell, 
  TableFooter, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '../table'

describe('Table Components', () => {
  describe('Table', () => {
    it('renders correctly', () => {
      render(
        <Table>
          <tbody>
            <tr>
              <td>Cell content</td>
            </tr>
          </tbody>
        </Table>
      )
      
      const table = screen.getByRole('table')
      expect(table).toBeInTheDocument()
      expect(table).toHaveClass('w-full', 'caption-bottom', 'text-sm')
    })

    it('wraps table in scrollable container', () => {
      render(
        <Table data-testid="table-wrapper">
          <tbody>
            <tr>
              <td>Content</td>
            </tr>
          </tbody>
        </Table>
      )
      
      const wrapper = screen.getByTestId('table-wrapper').parentElement
      expect(wrapper).toHaveClass('relative', 'w-full', 'overflow-auto')
    })
  })

  describe('TableHeader', () => {
    it('renders as thead element', () => {
      render(
        <table>
          <TableHeader>
            <tr>
              <th>Header</th>
            </tr>
          </TableHeader>
        </table>
      )
      
      const thead = screen.getByText('Header').closest('thead')
      expect(thead).toBeInTheDocument()
      expect(thead).toHaveClass('[&_tr]:border-b')
    })
  })

  describe('TableBody', () => {
    it('renders as tbody element', () => {
      render(
        <table>
          <TableBody>
            <tr>
              <td>Body content</td>
            </tr>
          </TableBody>
        </table>
      )
      
      const tbody = screen.getByText('Body content').closest('tbody')
      expect(tbody).toBeInTheDocument()
      expect(tbody).toHaveClass('[&_tr:last-child]:border-0')
    })
  })

  describe('TableFooter', () => {
    it('renders as tfoot element', () => {
      render(
        <table>
          <TableFooter>
            <tr>
              <td>Footer content</td>
            </tr>
          </TableFooter>
        </table>
      )
      
      const tfoot = screen.getByText('Footer content').closest('tfoot')
      expect(tfoot).toBeInTheDocument()
      expect(tfoot).toHaveClass('border-t', 'bg-muted/50', 'font-medium')
    })
  })

  describe('TableRow', () => {
    it('renders as tr element', () => {
      render(
        <table>
          <tbody>
            <TableRow>
              <td>Row content</td>
            </TableRow>
          </tbody>
        </table>
      )
      
      const row = screen.getByText('Row content').closest('tr')
      expect(row).toBeInTheDocument()
      expect(row).toHaveClass('border-b', 'transition-colors', 'hover:bg-muted/50')
    })
  })

  describe('TableHead', () => {
    it('renders as th element', () => {
      render(
        <table>
          <thead>
            <tr>
              <TableHead>Column Header</TableHead>
            </tr>
          </thead>
        </table>
      )
      
      const th = screen.getByText('Column Header')
      expect(th.tagName).toBe('TH')
      expect(th).toHaveClass('h-12', 'px-4', 'text-left', 'align-middle', 'font-medium', 'text-muted-foreground')
    })
  })

  describe('TableCell', () => {
    it('renders as td element', () => {
      render(
        <table>
          <tbody>
            <tr>
              <TableCell>Cell content</TableCell>
            </tr>
          </tbody>
        </table>
      )
      
      const td = screen.getByText('Cell content')
      expect(td.tagName).toBe('TD')
      expect(td).toHaveClass('p-4', 'align-middle')
    })
  })

  describe('TableCaption', () => {
    it('renders as caption element', () => {
      render(
        <Table>
          <TableCaption>Table caption</TableCaption>
          <tbody>
            <tr>
              <td>Content</td>
            </tr>
          </tbody>
        </Table>
      )
      
      const caption = screen.getByText('Table caption')
      expect(caption.tagName).toBe('CAPTION')
      expect(caption).toHaveClass('mt-4', 'text-sm', 'text-muted-foreground')
    })
  })

  describe('Complete Table', () => {
    it('renders full table structure', () => {
      render(
        <Table>
          <TableCaption>A list of your recent invoices.</TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead>Invoice</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Method</TableHead>
              <TableHead className="text-right">Amount</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell className="font-medium">INV001</TableCell>
              <TableCell>Paid</TableCell>
              <TableCell>Credit Card</TableCell>
              <TableCell className="text-right">$250.00</TableCell>
            </TableRow>
            <TableRow>
              <TableCell className="font-medium">INV002</TableCell>
              <TableCell>Pending</TableCell>
              <TableCell>PayPal</TableCell>
              <TableCell className="text-right">$150.00</TableCell>
            </TableRow>
          </TableBody>
          <TableFooter>
            <TableRow>
              <TableCell colSpan={3}>Total</TableCell>
              <TableCell className="text-right">$400.00</TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      )

      // Check structure
      expect(screen.getByRole('table')).toBeInTheDocument()
      expect(screen.getByText('A list of your recent invoices.')).toBeInTheDocument()
      
      // Check headers
      expect(screen.getByText('Invoice')).toBeInTheDocument()
      expect(screen.getByText('Status')).toBeInTheDocument()
      expect(screen.getByText('Method')).toBeInTheDocument()
      expect(screen.getByText('Amount')).toBeInTheDocument()
      
      // Check data
      expect(screen.getByText('INV001')).toBeInTheDocument()
      expect(screen.getByText('Paid')).toBeInTheDocument()
      expect(screen.getByText('Credit Card')).toBeInTheDocument()
      expect(screen.getByText('$250.00')).toBeInTheDocument()
      
      // Check footer
      expect(screen.getByText('Total')).toBeInTheDocument()
      expect(screen.getByText('$400.00')).toBeInTheDocument()
    })
  })
})

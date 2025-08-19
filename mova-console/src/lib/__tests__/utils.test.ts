import { cn } from '../utils'

describe('Utils', () => {
  describe('cn function', () => {
    it('merges class names correctly', () => {
      expect(cn('px-2 py-1', 'bg-red-500')).toBe('px-2 py-1 bg-red-500')
    })

    it('handles conditional classes', () => {
      expect(cn('base-class', true && 'conditional-class')).toBe('base-class conditional-class')
      expect(cn('base-class', false && 'conditional-class')).toBe('base-class')
    })

    it('handles undefined and null values', () => {
      expect(cn('base-class', undefined, null)).toBe('base-class')
    })

    it('handles arrays of classes', () => {
      expect(cn(['class1', 'class2'], 'class3')).toBe('class1 class2 class3')
    })

    it('handles objects with boolean values', () => {
      expect(cn({
        'active': true,
        'inactive': false,
        'default': true
      })).toBe('active default')
    })

    it('deduplicates conflicting Tailwind classes', () => {
      // This tests the tailwind-merge functionality
      expect(cn('px-2 px-4')).toBe('px-4')
      expect(cn('text-red-500 text-blue-500')).toBe('text-blue-500')
    })

    it('handles complex combinations', () => {
      const result = cn(
        'base-class',
        'px-2',
        { 'active': true, 'disabled': false },
        ['array-class1', 'array-class2'],
        'px-4' // Should override px-2
      )
      expect(result).toContain('base-class')
      expect(result).toContain('active')
      expect(result).toContain('array-class1')
      expect(result).toContain('array-class2')
      expect(result).toContain('px-4')
      expect(result).not.toContain('px-2')
      expect(result).not.toContain('disabled')
    })

    it('handles empty inputs', () => {
      expect(cn()).toBe('')
      expect(cn('')).toBe('')
      expect(cn([])).toBe('')
      expect(cn({})).toBe('')
    })

    it('trims whitespace correctly', () => {
      expect(cn('  class1  ', '  class2  ')).toBe('class1 class2')
    })

    it('handles multiple spaces', () => {
      expect(cn('class1    class2')).toBe('class1 class2')
    })

    it('works with variant combinations', () => {
      // Common pattern in UI components
      const variants = {
        size: {
          sm: 'text-sm px-2 py-1',
          md: 'text-base px-3 py-2',
          lg: 'text-lg px-4 py-3'
        },
        variant: {
          primary: 'bg-blue-500 text-white',
          secondary: 'bg-gray-200 text-gray-900'
        }
      }
      
      const result = cn(
        'base-button',
        variants.size.md,
        variants.variant.primary,
        'hover:opacity-80'
      )
      
      expect(result).toContain('base-button')
      expect(result).toContain('text-base')
      expect(result).toContain('px-3')
      expect(result).toContain('py-2')
      expect(result).toContain('bg-blue-500')
      expect(result).toContain('text-white')
      expect(result).toContain('hover:opacity-80')
    })

    it('handles responsive classes', () => {
      const result = cn(
        'text-sm',
        'md:text-base',
        'lg:text-lg',
        'xl:text-xl'
      )
      
      expect(result).toBe('text-sm md:text-base lg:text-lg xl:text-xl')
    })

    it('handles state modifiers', () => {
      const result = cn(
        'bg-blue-500',
        'hover:bg-blue-600',
        'focus:bg-blue-700',
        'active:bg-blue-800',
        'disabled:bg-gray-300'
      )
      
      expect(result).toContain('bg-blue-500')
      expect(result).toContain('hover:bg-blue-600')
      expect(result).toContain('focus:bg-blue-700')
      expect(result).toContain('active:bg-blue-800')
      expect(result).toContain('disabled:bg-gray-300')
    })

    it('handles dark mode classes', () => {
      const result = cn(
        'bg-white',
        'dark:bg-gray-900',
        'text-gray-900',
        'dark:text-white'
      )
      
      expect(result).toContain('bg-white')
      expect(result).toContain('dark:bg-gray-900')
      expect(result).toContain('text-gray-900')
      expect(result).toContain('dark:text-white')
    })
  })
})

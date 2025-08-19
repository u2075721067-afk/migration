import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import RetryPolicyEditor from '../RetryPolicyEditor';

// Mock fetch
global.fetch = jest.fn();

describe('RetryPolicyEditor', () => {
  beforeEach(() => {
    (fetch as jest.Mock).mockClear();
  });

  it('renders the retry policy editor', () => {
    render(<RetryPolicyEditor />);
    
    expect(screen.getByText('Retry Policy Editor')).toBeInTheDocument();
    expect(screen.getByText('Policy Name')).toBeInTheDocument();
    expect(screen.getByText('Description')).toBeInTheDocument();
    expect(screen.getByText('Retry Profile')).toBeInTheDocument();
  });

  it('shows all retry profiles in dropdown', () => {
    render(<RetryPolicyEditor />);
    
    const profileSelect = screen.getByDisplayValue(/balanced/);
    expect(profileSelect).toBeInTheDocument();
    
    // Check if all profiles are available as options
    const options = screen.getAllByRole('option');
    const profileNames = options.map(option => option.textContent);
    
    expect(profileNames.some(name => name?.includes('aggressive'))).toBe(true);
    expect(profileNames.some(name => name?.includes('balanced'))).toBe(true);
    expect(profileNames.some(name => name?.includes('conservative'))).toBe(true);
  });

  it('updates policy name when input changes', () => {
    render(<RetryPolicyEditor />);
    
    const nameInput = screen.getByPlaceholderText('Enter policy name');
    fireEvent.change(nameInput, { target: { value: 'Test Policy' } });
    
    expect(nameInput).toHaveValue('Test Policy');
  });

  it('updates description when textarea changes', () => {
    render(<RetryPolicyEditor />);
    
    const descriptionTextarea = screen.getByPlaceholderText('Describe the policy purpose');
    fireEvent.change(descriptionTextarea, { target: { value: 'Test description' } });
    
    expect(descriptionTextarea).toHaveValue('Test description');
  });

  it('changes retry profile when dropdown selection changes', () => {
    render(<RetryPolicyEditor />);
    
    const profileSelect = screen.getByDisplayValue(/balanced/);
    fireEvent.change(profileSelect, { target: { value: 'aggressive' } });
    
    // Check if the selected profile details are updated
    expect(screen.getByText('Selected Profile: aggressive')).toBeInTheDocument();
  });

  it('shows profile details for selected profile', () => {
    render(<RetryPolicyEditor />);
    
    // Should show balanced profile details by default
    expect(screen.getByText('Selected Profile: balanced')).toBeInTheDocument();
    expect(screen.getByText('Max Retries:')).toBeInTheDocument();
    expect(screen.getByText('Initial Delay:')).toBeInTheDocument();
    expect(screen.getByText('Max Delay:')).toBeInTheDocument();
    expect(screen.getByText('Backoff Multiplier:')).toBeInTheDocument();
  });

  it('adds a new condition when Add Condition button is clicked', () => {
    render(<RetryPolicyEditor />);
    
    const addButton = screen.getByText('Add Condition');
    fireEvent.click(addButton);
    
    expect(screen.getByText('Condition 1')).toBeInTheDocument();
    expect(screen.getByText('Error Type')).toBeInTheDocument();
    expect(screen.getByText('HTTP Status')).toBeInTheDocument();
  });

  it('removes condition when Remove button is clicked', () => {
    render(<RetryPolicyEditor />);
    
    // Add a condition first
    const addButton = screen.getByText('Add Condition');
    fireEvent.click(addButton);
    
    expect(screen.getByText('Condition 1')).toBeInTheDocument();
    
    // Remove the condition
    const removeButton = screen.getByText('Remove');
    fireEvent.click(removeButton);
    
    expect(screen.queryByText('Condition 1')).not.toBeInTheDocument();
  });

  it('updates condition fields when inputs change', () => {
    render(<RetryPolicyEditor />);
    
    // Add a condition
    const addButton = screen.getByText('Add Condition');
    fireEvent.click(addButton);
    
    // Update error type
    const errorTypeInput = screen.getByPlaceholderText('e.g., timeout, network');
    fireEvent.change(errorTypeInput, { target: { value: 'timeout' } });
    
    expect(errorTypeInput).toHaveValue('timeout');
    
    // Update HTTP status
    const httpStatusInput = screen.getByPlaceholderText('e.g., 408, 500');
    fireEvent.change(httpStatusInput, { target: { value: '408' } });
    
    expect(httpStatusInput).toHaveValue(408);
  });

  it('updates YAML preview when policy changes', async () => {
    render(<RetryPolicyEditor />);
    
    // Update policy name
    const nameInput = screen.getByPlaceholderText('Enter policy name');
    fireEvent.change(nameInput, { target: { value: 'Test Policy' } });
    
    // Check if YAML preview is updated
    await waitFor(() => {
      expect(screen.getByText('YAML Preview')).toBeInTheDocument();
      const yamlContent = screen.getByText(/name: Test Policy/);
      expect(yamlContent).toBeInTheDocument();
    });
  });

  it('disables Apply Policy button when required fields are missing', () => {
    render(<RetryPolicyEditor />);
    
    const applyButton = screen.getByText('Apply Policy');
    expect(applyButton).toBeDisabled();
    
    // Fill in name
    const nameInput = screen.getByPlaceholderText('Enter policy name');
    fireEvent.change(nameInput, { target: { value: 'Test Policy' } });
    
    // Still disabled without description
    expect(applyButton).toBeDisabled();
    
    // Fill in description
    const descriptionTextarea = screen.getByPlaceholderText('Describe the policy purpose');
    fireEvent.change(descriptionTextarea, { target: { value: 'Test description' } });
    
    // Now should be enabled
    expect(applyButton).not.toBeDisabled();
  });

  it('calls API when Apply Policy button is clicked', async () => {
    (fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ success: true }),
    });

    // Mock alert
    window.alert = jest.fn();

    render(<RetryPolicyEditor />);
    
    // Fill in required fields
    const nameInput = screen.getByPlaceholderText('Enter policy name');
    fireEvent.change(nameInput, { target: { value: 'Test Policy' } });
    
    const descriptionTextarea = screen.getByPlaceholderText('Describe the policy purpose');
    fireEvent.change(descriptionTextarea, { target: { value: 'Test description' } });
    
    // Click apply
    const applyButton = screen.getByText('Apply Policy');
    fireEvent.click(applyButton);
    
    await waitFor(() => {
      expect(fetch).toHaveBeenCalledWith('/api/v1/policies', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: 'Test Policy',
          description: 'Test description',
          retryProfile: 'balanced',
          conditions: [],
          enabled: true,
        }),
      });
    });

    expect(window.alert).toHaveBeenCalledWith('Policy applied successfully!');
  });

  it('shows error when API call fails', async () => {
    (fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      status: 400,
    });

    // Mock alert
    window.alert = jest.fn();

    render(<RetryPolicyEditor />);
    
    // Fill in required fields
    const nameInput = screen.getByPlaceholderText('Enter policy name');
    fireEvent.change(nameInput, { target: { value: 'Test Policy' } });
    
    const descriptionTextarea = screen.getByPlaceholderText('Describe the policy purpose');
    fireEvent.change(descriptionTextarea, { target: { value: 'Test description' } });
    
    // Click apply
    const applyButton = screen.getByText('Apply Policy');
    fireEvent.click(applyButton);
    
    await waitFor(() => {
      expect(window.alert).toHaveBeenCalledWith('Failed to apply policy');
    });
  });

  it('toggles enabled checkbox', () => {
    render(<RetryPolicyEditor />);
    
    const enabledCheckbox = screen.getByLabelText('Enable Policy') as HTMLInputElement;
    expect(enabledCheckbox.checked).toBe(true);
    
    fireEvent.click(enabledCheckbox);
    expect(enabledCheckbox.checked).toBe(false);
    
    fireEvent.click(enabledCheckbox);
    expect(enabledCheckbox.checked).toBe(true);
  });

  it('shows no conditions message when no conditions are defined', () => {
    render(<RetryPolicyEditor />);
    
    expect(screen.getByText('No conditions defined. Add conditions to make the policy more specific.')).toBeInTheDocument();
  });

  it('handles multiple conditions correctly', () => {
    render(<RetryPolicyEditor />);
    
    // Add two conditions
    const addButton = screen.getByText('Add Condition');
    fireEvent.click(addButton);
    fireEvent.click(addButton);
    
    expect(screen.getByText('Condition 1')).toBeInTheDocument();
    expect(screen.getByText('Condition 2')).toBeInTheDocument();
    
    // Remove first condition
    const removeButtons = screen.getAllByText('Remove');
    fireEvent.click(removeButtons[0]);
    
    // Should still have one condition (originally condition 2)
    expect(screen.getByText('Condition 1')).toBeInTheDocument();
    expect(screen.queryByText('Condition 2')).not.toBeInTheDocument();
  });
});

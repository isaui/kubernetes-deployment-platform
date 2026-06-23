/**
 * Utility functions for formatting values consistently
 */

/**
 * Converts byte-based values with different units (Ki, Mi, Gi) to a consistent format
 * This ensures values like "12.33Gi" and "32877912Ki" are displayed with the same unit
 * @param value The string value to convert (e.g., "12.33Gi", "32877912Ki")
 * @returns Formatted string with consistent unit
 */
export function formatBytesConsistently(value: string | number): string {
  if (typeof value === 'undefined' || value === null) {
    return '0 B';
  }
  
  // Convert to string if it's a number
  const strValue = String(value);
  
  // Extract the numeric part and the unit
  const match = strValue.match(/^([\d.]+)\s*([KMGTPEZY]?i?)B?$/i);
  
  if (!match) {
    // If no unit found or doesn't match pattern, return as is
    return strValue;
  }
  
  const numeric = parseFloat(match[1]);
  const unit = match[2]?.toUpperCase() || '';
  
  // Convert to bytes based on the unit
  let bytes = numeric;
  if (unit === 'KI' || unit === 'K') {
    bytes = numeric * 1024;
  } else if (unit === 'MI' || unit === 'M') {
    bytes = numeric * 1024 * 1024;
  } else if (unit === 'GI' || unit === 'G') {
    bytes = numeric * 1024 * 1024 * 1024;
  } else if (unit === 'TI' || unit === 'T') {
    bytes = numeric * 1024 * 1024 * 1024 * 1024;
  }
  
  // Convert back to a human-readable format with a consistent unit (Gi)
  if (bytes >= 1024 * 1024 * 1024) {
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GiB`;
  } else if (bytes >= 1024 * 1024) {
    return `${(bytes / (1024 * 1024)).toFixed(2)} MiB`;
  } else if (bytes >= 1024) {
    return `${(bytes / 1024).toFixed(2)} KiB`;
  } else {
    return `${bytes.toFixed(2)} B`;
  }
}

/**
 * Formats an ISO date string to a more human-readable format
 * @param dateString ISO date string (e.g., "2023-06-15T10:30:00Z")
 * @returns Formatted date string (e.g., "Jun 15, 2023, 10:30 AM")
 */
export function formatDate(dateString: string): string {
  if (!dateString) return '-';
  
  try {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric', 
      month: 'short', 
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    }).format(date);
  } catch (error) {
    // Return as-is if there's an error parsing the date
    return dateString;
  }
}

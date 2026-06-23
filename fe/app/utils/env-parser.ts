// utils/env-parser.ts - Utility functions for environment variable parsing

export interface ParseResult {
  variables: Record<string, string>;
  errors: string[];
  warnings: string[];
}

/**
 * Validates environment variable name according to common conventions
 */
export function validateEnvVarName(name: string): { valid: boolean; error?: string } {
  if (!name) {
    return { valid: false, error: 'Variable name cannot be empty' };
  }
  
  if (!/^[A-Z0-9_]+$/i.test(name)) {
    return { 
      valid: false, 
      error: 'Variable name can only contain letters, numbers, and underscores' 
    };
  }
  
  if (name.startsWith('_')) {
    return { 
      valid: false, 
      error: 'Variable name should not start with underscore' 
    };
  }
  
  if (!/^[A-Z]/.test(name)) {
    return { 
      valid: false, 
      error: 'Variable name should start with a letter' 
    };
  }
  
  return { valid: true };
}

/**
 * Parses .env format content
 */
export function parseEnvContent(content: string): ParseResult {
  const variables: Record<string, string> = {};
  const errors: string[] = [];
  const warnings: string[] = [];
  
  if (!content.trim()) {
    return { variables, errors, warnings };
  }
  
  const lines = content.split('\n');
  let currentLine = '';
  let isMultiline = false;
  let multilineKey = '';
  let multilineValue = '';
  
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const lineNumber = i + 1;
    
    // Handle multiline continuation
    if (isMultiline) {
      multilineValue += '\n' + line;
      
      // Check if multiline ends
      if ((multilineValue.includes('"') && multilineValue.match(/"$/)) ||
          (multilineValue.includes("'") && multilineValue.match(/'$/))) {
        // Process the complete multiline variable
        const result = processEnvLine(multilineKey + '=' + multilineValue, lineNumber);
        if (result.error) {
          errors.push(result.error);
        } else if (result.key && result.value !== undefined) {
          variables[result.key] = result.value;
        }
        
        isMultiline = false;
        multilineKey = '';
        multilineValue = '';
      }
      continue;
    }
    
    const trimmedLine = line.trim();
    
    // Skip empty lines and comments
    if (!trimmedLine || trimmedLine.startsWith('#')) {
      continue;
    }
    
    // Check for multiline start
    const equalIndex = trimmedLine.indexOf('=');
    if (equalIndex !== -1) {
      const key = trimmedLine.substring(0, equalIndex).trim();
      const value = trimmedLine.substring(equalIndex + 1).trim();
      
      // Check if this starts a multiline value
      if ((value.startsWith('"') && !value.match(/^".*"$/)) ||
          (value.startsWith("'") && !value.match(/^'.*'$/))) {
        isMultiline = true;
        multilineKey = key;
        multilineValue = value;
        continue;
      }
    }
    
    // Process single line
    const result = processEnvLine(trimmedLine, lineNumber);
    if (result.error) {
      errors.push(result.error);
    } else if (result.warning) {
      warnings.push(result.warning);
    }
    
    if (result.key && result.value !== undefined) {
      // Check for duplicates
      if (variables[result.key]) {
        warnings.push(`Line ${lineNumber}: Duplicate variable "${result.key}" (previous value will be overwritten)`);
      }
      variables[result.key] = result.value;
    }
  }
  
  // Check if we ended with an incomplete multiline
  if (isMultiline) {
    errors.push(`Incomplete multiline value for variable "${multilineKey}"`);
  }
  
  return { variables, errors, warnings };
}

/**
 * Processes a single line from .env format
 */
function processEnvLine(line: string, lineNumber: number): {
  key?: string;
  value?: string;
  error?: string;
  warning?: string;
} {
  const equalIndex = line.indexOf('=');
  
  if (equalIndex === -1) {
    return { error: `Line ${lineNumber}: Invalid format - missing '=' sign` };
  }
  
  if (equalIndex === 0) {
    return { error: `Line ${lineNumber}: Missing variable name before '='` };
  }
  
  const key = line.substring(0, equalIndex).trim();
  let value = line.substring(equalIndex + 1).trim();
  
  // Validate key
  const keyValidation = validateEnvVarName(key);
  if (!keyValidation.valid) {
    return { error: `Line ${lineNumber}: ${keyValidation.error}` };
  }
  
  // Process value - remove quotes if present
  if ((value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))) {
    value = value.slice(1, -1);
    
    // Unescape quotes in double-quoted strings
    if (line.substring(equalIndex + 1).trim().startsWith('"')) {
      value = value.replace(/\\"/g, '"').replace(/\\\\/g, '\\');
    }
  }
  
  // Check for potentially problematic values
  let warning: string | undefined;
  if (value.includes('\n') && !line.includes('"') && !line.includes("'")) {
    warning = `Line ${lineNumber}: Multiline value detected but not properly quoted`;
  } else if (value.includes(' ') && !line.includes('"') && !line.includes("'")) {
    warning = `Line ${lineNumber}: Value contains spaces but is not quoted`;
  }
  
  return { key, value, warning };
}

/**
 * Parses JSON format content
 */
export function parseJsonContent(content: string): ParseResult {
  const variables: Record<string, string> = {};
  const errors: string[] = [];
  const warnings: string[] = [];
  
  if (!content.trim()) {
    return { variables, errors, warnings };
  }
  
  try {
    const parsed = JSON.parse(content);
    
    if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
      errors.push('JSON must be an object with key-value pairs');
      return { variables, errors, warnings };
    }
    
    for (const [key, value] of Object.entries(parsed)) {
      // Validate key
      const keyValidation = validateEnvVarName(key);
      if (!keyValidation.valid) {
        errors.push(`Variable "${key}": ${keyValidation.error}`);
        continue;
      }
      
      // Convert value to string and validate
      if (value === null || value === undefined) {
        warnings.push(`Variable "${key}": null/undefined value converted to empty string`);
        variables[key] = '';
      } else if (typeof value === 'object') {
        errors.push(`Variable "${key}": object values are not supported`);
        continue;
      } else {
        variables[key] = String(value);
        
        // Warn about non-string values
        if (typeof value !== 'string') {
          warnings.push(`Variable "${key}": ${typeof value} value converted to string`);
        }
      }
    }
  } catch (err) {
    errors.push(`Invalid JSON: ${err instanceof Error ? err.message : 'Unknown error'}`);
  }
  
  return { variables, errors, warnings };
}

/**
 * Converts environment variables to .env format
 */
export function variablesToEnvFormat(variables: Record<string, string>): string {
  return Object.entries(variables)
    .map(([key, value]) => {
      // Quote values that contain spaces, special characters, or newlines
      if (value.includes(' ') || 
          value.includes('\n') || 
          value.includes('\t') ||
          /[<>|&;$`"']/.test(value)) {
        const escapedValue = value.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
        return `${key}="${escapedValue}"`;
      }
      return `${key}=${value}`;
    })
    .join('\n');
}

/**
 * Converts environment variables to JSON format
 */
export function variablesToJsonFormat(variables: Record<string, string>): string {
  return JSON.stringify(variables, null, 2);
}

/**
 * Detects the format of the input content
 */
export function detectFormat(content: string): 'env' | 'json' | 'unknown' {
  const trimmed = content.trim();
  
  if (!trimmed) {
    return 'unknown';
  }
  
  // Check if it looks like JSON
  if ((trimmed.startsWith('{') && trimmed.endsWith('}')) ||
      (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
    try {
      JSON.parse(trimmed);
      return 'json';
    } catch {
      // Not valid JSON, might be .env
    }
  }
  
  // Check if it looks like .env format
  const lines = trimmed.split('\n');
  let envLikeLines = 0;
  let totalLines = 0;
  
  for (const line of lines) {
    const trimmedLine = line.trim();
    if (trimmedLine && !trimmedLine.startsWith('#')) {
      totalLines++;
      if (trimmedLine.includes('=')) {
        envLikeLines++;
      }
    }
  }
  
  // If most non-comment lines have = signs, it's probably .env format
  if (totalLines > 0 && envLikeLines / totalLines > 0.5) {
    return 'env';
  }
  
  return 'unknown';
}

/**
 * Validates a complete set of environment variables
 */
export function validateEnvironmentVariables(
  variables: Record<string, string>
): { valid: boolean; errors: string[]; warnings: string[] } {
  const errors: string[] = [];
  const warnings: string[] = [];
  
  // Check for reserved/system variables
  const reservedNames = [
    'PATH', 'HOME', 'USER', 'SHELL', 'PWD', 'TERM',
    'KUBERNETES_SERVICE_HOST', 'KUBERNETES_SERVICE_PORT'
  ];
  
  // Check for common naming patterns that might be problematic
  const problematicPatterns = [
    { pattern: /^[0-9]/, message: 'starts with a number' },
    { pattern: /[a-z]/, message: 'contains lowercase letters (consider using uppercase)' },
    { pattern: /[-]/, message: 'contains hyphens (use underscores instead)' },
  ];
  
  for (const [key, value] of Object.entries(variables)) {
    // Check reserved names
    if (reservedNames.includes(key)) {
      warnings.push(`Variable "${key}" is a system variable and might be overridden`);
    }
    
    // Check naming patterns
    for (const { pattern, message } of problematicPatterns) {
      if (pattern.test(key)) {
        warnings.push(`Variable "${key}" ${message}`);
        break;
      }
    }
    
    // Check value length
    if (value.length > 1000) {
      warnings.push(`Variable "${key}" has a very long value (${value.length} characters)`);
    }
    
    // Check for potential secrets in value
    if (/password|secret|key|token/i.test(key) && value.length < 8) {
      warnings.push(`Variable "${key}" appears to be a secret but has a short value`);
    }
  }
  
  return {
    valid: errors.length === 0,
    errors,
    warnings
  };
}
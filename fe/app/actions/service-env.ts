import type { Service } from '~/types/service';

/**
 * Get all environment variables for a service
 * 
 * @param {string} serviceId - Service ID to get environment variables for
 * @returns Promise with environment variables as key-value pairs
 */
export async function getServiceEnvVars(serviceId: string): Promise<Record<string, string>> {
  const response = await fetch(
    `/api/v1/services/${serviceId}`,
    {
      method: "GET",
      credentials: "include",
      headers: {
        "Accept": "application/json",
      },
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.error || "Failed to fetch service"
    );
  }

  const result = await response.json();
  return result.data.envVars || {};
}

/**
 * Set a single environment variable for a service
 * 
 * @param {string} serviceId - Service ID to add/update environment variable for
 * @param {string} name - Environment variable name
 * @param {string} value - Environment variable value
 * @returns Promise with updated environment variables
 */
export async function setServiceEnvVar(
  serviceId: string,
  name: string,
  value: string
): Promise<Record<string, string>> {
  // First get the current service
  const response = await fetch(
    `/api/v1/services/${serviceId}`,
    {
      method: "GET",
      credentials: "include",
      headers: {
        "Accept": "application/json",
      },
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.error || "Failed to fetch service"
    );
  }

  const service = await response.json();
  const serviceData = service.data;
  
  // Update the envVars property
  const updatedEnvVars = { 
    ...serviceData.envVars || {}, 
    [name]: value 
  };
  
  // Update the service with the new envVars
  const updateResponse = await fetch(
    `/api/v1/services/${serviceId}`,
    {
      method: "PUT",
      credentials: "include",
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        ...serviceData,
        envVars: updatedEnvVars
      }),
    }
  );

  if (!updateResponse.ok) {
    const errorData = await updateResponse.json().catch(() => ({}));
    throw new Error(
      errorData.error || "Failed to set environment variable"
    );
  }

  const updateResult = await updateResponse.json();
  return updateResult.data.envVars || {};
}

/**
 * Delete an environment variable from a service
 * 
 * @param {string} serviceId - Service ID to delete environment variable from
 * @param {string} name - Environment variable name to delete
 * @returns Promise with updated environment variables
 */
export async function deleteServiceEnvVar(
  serviceId: string,
  name: string
): Promise<Record<string, string>> {
  // First get the current service
  const response = await fetch(
    `/api/v1/services/${serviceId}`,
    {
      method: "GET",
      credentials: "include",
      headers: {
        "Accept": "application/json",
      },
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.error || "Failed to fetch service"
    );
  }

  const service = await response.json();
  const serviceData = service.data;
  
  // Create a copy of envVars and delete the specified variable
  const updatedEnvVars = { ...serviceData.envVars || {} };
  delete updatedEnvVars[name];
  
  // Update the service with the modified envVars
  const updateResponse = await fetch(
    `/api/v1/services/${serviceId}`,
    {
      method: "PUT",
      credentials: "include",
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        ...serviceData,
        envVars: updatedEnvVars
      }),
    }
  );

  if (!updateResponse.ok) {
    const errorData = await updateResponse.json().catch(() => ({}));
    throw new Error(
      errorData.error || "Failed to delete environment variable"
    );
  }

  const updateResult = await updateResponse.json();
  return updateResult.data.envVars || {};
}

/**
 * Update multiple environment variables for a service at once
 * 
 * @param {string} serviceId - Service ID to update environment variables for
 * @param {Record<string, string>} envVars - Complete set of environment variables
 * @returns Promise with updated environment variables
 */
export async function updateServiceEnvVars(
  serviceId: string,
  envVars: Record<string, string>
): Promise<Record<string, string>> {
  // First get the current service
  const response = await fetch(
    `/api/v1/services/${serviceId}`,
    {
      method: "GET",
      credentials: "include",
      headers: {
        "Accept": "application/json",
      },
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.error || "Failed to fetch service"
    );
  }

  const service = await response.json();
  const serviceData = service.data;
  
  // Update the service with the new envVars
  const updateResponse = await fetch(
    `/api/v1/services/${serviceId}`,
    {
      method: "PUT",
      credentials: "include",
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        ...serviceData,
        envVars: envVars
      }),
    }
  );

  if (!updateResponse.ok) {
    const errorData = await updateResponse.json().catch(() => ({}));
    throw new Error(
      errorData.error || "Failed to update environment variables"
    );
  }

  const updateResult = await updateResponse.json();
  return updateResult.data.envVars || {};
}

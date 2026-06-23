import { Environment } from "~/types/project";

// Function to list environments for a project
export async function listProjectEnvironments(
  projectId: string
): Promise<Environment[]> {
  const response = await fetch(
    `/api/v1/projects/${projectId}/environments`,
    {
      method: "GET",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    }
  );

  if (!response.ok) {
    throw new Error("Failed to fetch environments");
  }

  const result = await response.json();
  return result.data.environments;
}

// Function to get a specific environment
export async function getEnvironment(
  environmentId: string
): Promise<Environment> {
  const response = await fetch(
    `/api/v1/environments/${environmentId}`,
    {
      method: "GET",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    }
  );

  if (!response.ok) {
    throw new Error("Failed to fetch environment");
  }

  const result = await response.json();
  return result.data;
}

// Function to create a new environment
export async function createEnvironment(
  data: { name: string; description: string; projectId: string }
): Promise<Environment> {
  const response = await fetch(
    `/api/v1/environments`,
    {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.error || "Failed to create environment"
    );
  }

  const result = await response.json();
  return result.data;
}

// Function to update an environment
export async function updateEnvironment(
  environmentId: string,
  data: { name: string; description: string; projectId?: string }
): Promise<Environment> {
  const response = await fetch(
    `/api/v1/environments/${environmentId}`,
    {
      method: "PUT",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.error || "Failed to update environment"
    );
  }

  const result = await response.json();
  return result.data;
}

// Function to delete an environment
export async function deleteEnvironment(
  environmentId: string
): Promise<void> {
  const response = await fetch(
    `/api/v1/environments/${environmentId}`,
    {
      method: "DELETE",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.error || "Failed to delete environment"
    );
  }
}

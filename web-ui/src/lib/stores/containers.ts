import { writable, derived, get } from "svelte/store";
import { token } from "./auth";

// Types
export interface Container {
  id: string;
  db_id?: string;
  user_id: string;
  name: string;
  image: string;
  status: "running" | "stopped" | "creating" | "error";
  created_at: string;
  last_used_at?: string;
  idle_seconds?: number;
  ip_address?: string;
}

export interface CreatingContainer {
  name: string;
  image: string;
  progress: number;
  message: string;
  stage: string;
}

export interface ContainersState {
  containers: Container[];
  isLoading: boolean;
  error: string | null;
  limit: number;
  creating: CreatingContainer | null;
}

// Initial state
const initialState: ContainersState = {
  containers: [],
  isLoading: false,
  error: null,
  limit: 2,
  creating: null,
};

// Helper to get auth token
function getToken(): string | null {
  return get(token);
}

// Helper for API calls
async function apiCall<T>(
  endpoint: string,
  options: RequestInit = {},
): Promise<{ data?: T; error?: string; status: number }> {
  const authToken = getToken();
  if (!authToken) {
    return { error: "Not authenticated", status: 401 };
  }

  try {
    const response = await fetch(endpoint, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
        ...options.headers,
      },
    });

    const data = await response.json().catch(() => ({}));

    if (!response.ok) {
      return {
        error: data.error || `Request failed with status ${response.status}`,
        status: response.status,
      };
    }

    return { data, status: response.status };
  } catch (e) {
    return {
      error: e instanceof Error ? e.message : "Network error",
      status: 0,
    };
  }
}

// Create the store
function createContainersStore() {
  const { subscribe, set, update } = writable<ContainersState>(initialState);

  return {
    subscribe,

    // Reset store
    reset() {
      set(initialState);
    },

    // Fetch all containers
    async fetchContainers() {
      update((state) => ({ ...state, isLoading: true, error: null }));

      const { data, error } = await apiCall<{
        containers: Container[];
        count: number;
        limit: number;
      }>("/api/containers");

      if (error) {
        update((state) => ({ ...state, isLoading: false, error }));
        return { success: false, error };
      }

      update((state) => ({
        ...state,
        containers: data?.containers || [],
        limit: data?.limit || 2,
        isLoading: false,
        error: null,
      }));

      return { success: true, containers: data?.containers || [] };
    },

    // Get a single container
    async getContainer(id: string) {
      const { data, error } = await apiCall<Container>(`/api/containers/${id}`);

      if (error) {
        return { success: false, error };
      }

      // Update container in store if it exists
      update((state) => ({
        ...state,
        containers: state.containers.map((c) =>
          c.id === id ? { ...c, ...data } : c,
        ),
      }));

      return { success: true, container: data };
    },

    // Create a new container
    async createContainer(name: string, image: string, customImage?: string) {
      update((state) => ({ ...state, isLoading: true, error: null }));

      const body: Record<string, string> = { name, image };
      if (image === "custom" && customImage) {
        body.custom_image = customImage;
      }

      const { data, error, status } = await apiCall<Container>(
        "/api/containers",
        {
          method: "POST",
          body: JSON.stringify(body),
        },
      );

      if (error) {
        update((state) => ({ ...state, isLoading: false, error }));
        return { success: false, error, status };
      }

      // Add new container to store
      update((state) => ({
        ...state,
        containers: [data!, ...state.containers],
        isLoading: false,
        error: null,
      }));

      return { success: true, container: data };
    },

    // Create container with progress
    // Uses regular endpoint since SSE doesn't work well with Cloudflare/PipeOps proxies
    createContainerWithProgress(
      name: string,
      image: string,
      customImage?: string,
      onProgress?: (event: ProgressEvent) => void,
      onComplete?: (container: Container) => void,
      onError?: (error: string) => void,
    ) {
      // Use the regular (non-SSE) endpoint directly since SSE is blocked by Cloudflare
      this.createContainerFallback(
        name,
        image,
        customImage,
        onProgress,
        onComplete,
        onError,
      );
    },

    // Fallback container creation without SSE (for platforms that don't support streaming)
    async createContainerFallback(
      name: string,
      image: string,
      customImage?: string,
      onProgress?: (event: ProgressEvent) => void,
      onComplete?: (container: Container) => void,
      onError?: (error: string) => void,
    ) {
      const authToken = getToken();
      if (!authToken) {
        onError?.("Not authenticated");
        return;
      }

      // Update progress to show we're using fallback
      update((state) => ({
        ...state,
        creating: state.creating
          ? {
              ...state.creating,
              progress: 10,
              message: "Creating container...",
              stage: "creating",
            }
          : null,
      }));

      onProgress?.({
        stage: "creating",
        message: "Creating container (this may take a moment)...",
        progress: 10,
      });

      const body: Record<string, string> = { name, image };
      if (image === "custom" && customImage) {
        body.custom_image = customImage;
      }

      try {
        const response = await fetch("/api/containers", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${authToken}`,
          },
          body: JSON.stringify(body),
        });

        const data = await response.json();

        if (!response.ok) {
          update((state) => ({ ...state, creating: null }));
          onError?.(data.error || "Failed to create container");
          return;
        }

        const container: Container = {
          id: data.id,
          db_id: data.db_id,
          user_id: data.user_id,
          name: data.name,
          image: data.image,
          status: data.status || "running",
          created_at: data.created_at,
          ip_address: data.ip_address,
        };

        update((state) => ({
          ...state,
          containers: [container, ...state.containers],
          creating: null,
        }));

        onProgress?.({
          stage: "ready",
          message: "Terminal ready!",
          progress: 100,
          complete: true,
          container_id: container.id,
        });

        onComplete?.(container);
      } catch (e) {
        update((state) => ({ ...state, creating: null }));
        onError?.(
          e instanceof Error ? e.message : "Failed to create container",
        );
      }
    },

    // Start a container
    async startContainer(id: string) {
      update((state) => ({
        ...state,
        containers: state.containers.map((c) =>
          c.id === id ? { ...c, status: "creating" as const } : c,
        ),
      }));

      const { data, error } = await apiCall<{
        id: string;
        status: string;
        recreated?: boolean;
      }>(`/api/containers/${id}/start`, { method: "POST" });

      if (error) {
        update((state) => ({
          ...state,
          containers: state.containers.map((c) =>
            c.id === id ? { ...c, status: "stopped" as const } : c,
          ),
        }));
        return { success: false, error };
      }

      // Handle recreated container (new ID)
      if (data?.recreated && data.id !== id) {
        update((state) => ({
          ...state,
          containers: state.containers.map((c) =>
            c.id === id ? { ...c, id: data.id, status: "running" as const } : c,
          ),
        }));
        return { success: true, newId: data.id, recreated: true };
      }

      update((state) => ({
        ...state,
        containers: state.containers.map((c) =>
          c.id === id ? { ...c, status: "running" as const } : c,
        ),
      }));

      return { success: true };
    },

    // Stop a container
    async stopContainer(id: string) {
      update((state) => ({
        ...state,
        containers: state.containers.map((c) =>
          c.id === id ? { ...c, status: "creating" as const } : c,
        ),
      }));

      const { error } = await apiCall(`/api/containers/${id}/stop`, {
        method: "POST",
      });

      if (error) {
        update((state) => ({
          ...state,
          containers: state.containers.map((c) =>
            c.id === id ? { ...c, status: "running" as const } : c,
          ),
        }));
        return { success: false, error };
      }

      update((state) => ({
        ...state,
        containers: state.containers.map((c) =>
          c.id === id ? { ...c, status: "stopped" as const } : c,
        ),
      }));

      return { success: true };
    },

    // Delete a container
    async deleteContainer(id: string) {
      const { error } = await apiCall(`/api/containers/${id}`, {
        method: "DELETE",
      });

      if (error) {
        return { success: false, error };
      }

      update((state) => ({
        ...state,
        containers: state.containers.filter((c) => c.id !== id),
      }));

      return { success: true };
    },

    // Update container status locally
    updateStatus(id: string, status: Container["status"]) {
      update((state) => ({
        ...state,
        containers: state.containers.map((c) =>
          c.id === id ? { ...c, status } : c,
        ),
      }));
    },
  };
}

// Progress event type
export interface ProgressEvent {
  stage: string;
  message: string;
  progress: number;
  detail?: string;
  error?: string;
  complete?: boolean;
  container_id?: string;
}

// Export the store
export const containers = createContainersStore();

// Derived stores
export const runningContainers = derived(containers, ($containers) =>
  $containers.containers.filter((c) => c.status === "running"),
);

export const stoppedContainers = derived(containers, ($containers) =>
  $containers.containers.filter((c) => c.status === "stopped"),
);

export const containerCount = derived(
  containers,
  ($containers) => $containers.containers.length,
);

export const isAtLimit = derived(
  containers,
  ($containers) => $containers.containers.length >= $containers.limit,
);

export const isCreating = derived(
  containers,
  ($containers) => $containers.creating !== null,
);

export const creatingContainer = derived(
  containers,
  ($containers) => $containers.creating,
);

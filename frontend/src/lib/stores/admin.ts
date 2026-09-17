import { get, writable } from "svelte/store";
import { api, getWebSocketUrl } from "../utils/api";
import { createRexecWebSocket } from "../utils/ws";
import type { User } from "./auth";
import type { Container } from "./containers";

// Types
export interface AdminUser extends User {
  created_at: string;
  updated_at?: string;
  containerCount: number;
}

export interface AdminContainer extends Container {
  username: string; // Owner's username
  userEmail: string; // Owner's email
}

export interface AdminTerminal {
  id: string;
  containerId: string;
  name: string;
  status: "connected" | "disconnected" | "error";
  userId: string;
  username: string;
  connectedAt: string;
}

export interface AdminAgent {
  id: string;
  user_id: string;
  username: string;
  name: string;
  description?: string;
  os: string;
  arch: string;
  shell: string;
  distro?: string;
  status: string;
  created_at: string;
  last_ping?: string;
  system_info?: any;
}

export interface AdminUsageTotals {
  users: number;
  subscribers: number;
  containers: number;
  activeSessions: number;
  logins: number;
  agents: number;
  onlineAgents: number;
  recordings: number;
  recordingHours: number;
}

export interface AdminUserList {
  users: AdminUser[];
  page: number;
  perPage: number;
  total: number;
  totalPages: number;
}

export interface AdminListQuery {
  page: number;
  perPage: number;
  search: string;
}

export interface AdminUsageActivity {
  newUsers: number;
  newContainers: number;
  newSessions: number;
  newLogins: number;
  newAgents: number;
  newRecordings: number;
  recordingHours: number;
}

export interface AdminUsagePoint {
  bucketStart: string;
  bucketLabel: string;
  newUsers: number;
  newContainers: number;
  newSessions: number;
  newLogins: number;
  newAgents: number;
  newRecordings: number;
}

export interface AdminUsageStats {
  range: string;
  interval: "hour" | "day" | "week" | "month";
  from: string;
  to: string;
  totals: AdminUsageTotals;
  activity: AdminUsageActivity;
  timeline: AdminUsagePoint[];
}

// Backend AdminEvent interface
export interface AdminEvent<T = any> {
  type:
    | "user_created"
    | "user_updated"
    | "user_deleted"
    | "container_created"
    | "container_updated"
    | "container_deleted"
    | "session_created"
    | "session_updated"
    | "session_deleted";
  payload: T;
  timestamp: string; // ISO 8601 string
}

export interface AdminState {
  users: AdminUser[];
  usersQuery: AdminListQuery;
  usersTotal: number;
  usersTotalPages: number;
  usersLoading: boolean;
  usersLoaded: boolean;
  subscribers: AdminUser[];
  subscribersQuery: AdminListQuery;
  subscribersTotal: number;
  subscribersTotalPages: number;
  subscribersLoading: boolean;
  subscribersLoaded: boolean;
  containers: AdminContainer[];
  containersLoading: boolean;
  containersLoaded: boolean;
  terminals: AdminTerminal[];
  terminalsLoading: boolean;
  terminalsLoaded: boolean;
  agents: AdminAgent[];
  agentsLoading: boolean;
  agentsLoaded: boolean;
  stats: AdminUsageStats | null;
  statsLoading: boolean;
  statsLoaded: boolean;
  isLoading: boolean;
  error: string | null;
  ws: WebSocket | null;
  wsConnected: boolean;
  wsReconnectAttempts: number;
  wsMaxReconnectAttempts: number;
  wsReconnectInterval: number; // in milliseconds
}

const defaultListQuery: AdminListQuery = {
  page: 1,
  perPage: 25,
  search: "",
};

const initialState: AdminState = {
  users: [],
  usersQuery: { ...defaultListQuery },
  usersTotal: 0,
  usersTotalPages: 0,
  usersLoading: false,
  usersLoaded: false,
  subscribers: [],
  subscribersQuery: { ...defaultListQuery },
  subscribersTotal: 0,
  subscribersTotalPages: 0,
  subscribersLoading: false,
  subscribersLoaded: false,
  containers: [],
  containersLoading: false,
  containersLoaded: false,
  terminals: [],
  terminalsLoading: false,
  terminalsLoaded: false,
  agents: [],
  agentsLoading: false,
  agentsLoaded: false,
  stats: null,
  statsLoading: false,
  statsLoaded: false,
  isLoading: false,
  error: null,
  ws: null,
  wsConnected: false,
  wsReconnectAttempts: 0,
  wsMaxReconnectAttempts: 10,
  wsReconnectInterval: 1000, // 1 second
};

function anyLoading(state: AdminState): boolean {
  return (
    state.statsLoading ||
    state.usersLoading ||
    state.subscribersLoading ||
    state.containersLoading ||
    state.terminalsLoading ||
    state.agentsLoading
  );
}

function createAdminStore() {
  const store = writable<AdminState>(initialState);
  const { subscribe, update } = store;

  let ws: WebSocket | null = null;
  let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
  let usersRefetchTimer: ReturnType<typeof setTimeout> | null = null;

  function scheduleUsersRefetch() {
    if (usersRefetchTimer) clearTimeout(usersRefetchTimer);
    usersRefetchTimer = setTimeout(() => {
      const state = get(store);
      if (state.usersLoaded) {
        void fetchUsersPage(false);
      }
      if (state.subscribersLoaded) {
        void fetchUsersPage(true);
      }
    }, 250);
  }

  async function connectWebSocket() {
    update((state) => ({
      ...state,
      wsConnected: false,
      error: null,
    }));

    if (typeof window === "undefined") return;

    const token = localStorage.getItem("rexec_token"); // Assuming token is stored here
    if (!token) {
      console.warn("Admin WebSocket: No authentication token found.");
      update((state) => ({ ...state, error: "Authentication token missing." }));
      return;
    }

    const wsUrl = getWebSocketUrl("/ws/admin/events");

    ws = createRexecWebSocket(wsUrl, token);
    update((state) => ({ ...state, ws }));

    ws.onopen = () => {
      console.log("Admin WebSocket: Connected");
      update((state) => ({
        ...state,
        wsConnected: true,
        wsReconnectAttempts: 0,
        wsReconnectInterval: initialState.wsReconnectInterval,
      }));
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
      }
    };

    ws.onmessage = (event) => {
      const adminEvent: AdminEvent = JSON.parse(event.data);
      if (
        adminEvent.type === "user_created" ||
        adminEvent.type === "user_updated" ||
        adminEvent.type === "user_deleted"
      ) {
        scheduleUsersRefetch();
        return;
      }

      update((state) => {
        let newContainers = [...state.containers];
        let newTerminals = [...state.terminals];

        switch (adminEvent.type) {
          case "container_created":
            newContainers = [
              ...newContainers,
              adminEvent.payload as AdminContainer,
            ];
            break;
          case "container_updated":
            newContainers = newContainers.map((c) =>
              c.id === (adminEvent.payload as AdminContainer).id
                ? (adminEvent.payload as AdminContainer)
                : c,
            );
            break;
          case "container_deleted":
            newContainers = newContainers.filter(
              (c) => c.id !== (adminEvent.payload as AdminContainer).id,
            );
            break;
          case "session_created":
            newTerminals = [
              ...newTerminals,
              adminEvent.payload as AdminTerminal,
            ];
            break;
          case "session_updated":
            newTerminals = newTerminals.map((t) =>
              t.id === (adminEvent.payload as AdminTerminal).id
                ? (adminEvent.payload as AdminTerminal)
                : t,
            );
            break;
          case "session_deleted":
            newTerminals = newTerminals.filter(
              (t) => t.id !== (adminEvent.payload as AdminTerminal).id,
            );
            break;
          default:
            console.warn(
              "Admin WebSocket: Unknown event type",
              adminEvent.type,
            );
        }

        return {
          ...state,
          containers: newContainers,
          terminals: newTerminals,
        };
      });
    };

    ws.onclose = (event) => {
      console.log(
        `Admin WebSocket: Disconnected (Code: ${event.code}, Reason: ${event.reason})`,
      );
      update((state) => ({ ...state, wsConnected: false }));

      if (event.code !== 1000 && event.code !== 1001) {
        // Don't try to reconnect on normal closures (1000: Normal, 1001: Going Away)
        reconnect();
      }
    };

    ws.onerror = (error) => {
      console.error("Admin WebSocket: Error", error);
      update((state) => ({ ...state, error: "WebSocket error" }));
      ws?.close(); // Close to trigger onclose and reconnect logic
    };
  }

  function reconnect() {
    update((state) => {
      if (state.wsReconnectAttempts < state.wsMaxReconnectAttempts) {
        const newReconnectAttempts = state.wsReconnectAttempts + 1;
        const newReconnectInterval = state.wsReconnectInterval * 2; // Exponential backoff
        reconnectTimeout = setTimeout(
          connectWebSocket,
          newReconnectInterval,
        ) as ReturnType<typeof setTimeout>;
        console.log(
          `Admin WebSocket: Reconnecting in ${
            newReconnectInterval / 1000
          }s (Attempt ${newReconnectAttempts})`,
        );
        return {
          ...state,
          wsReconnectAttempts: newReconnectAttempts,
          wsReconnectInterval: newReconnectInterval,
        };
      } else {
        console.error(
          "Admin WebSocket: Max reconnect attempts reached. Please refresh.",
        );
        return { ...state, error: "Max reconnect attempts reached." };
      }
    });
  }

  function disconnectWebSocket() {
    if (ws) {
      ws.close(1000, "Component unmounted"); // Normal closure
      ws = null;
    }
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout);
      reconnectTimeout = null;
    }
    if (usersRefetchTimer) {
      clearTimeout(usersRefetchTimer);
      usersRefetchTimer = null;
    }
    update((state) => ({ ...state, wsConnected: false, ws: null }));
  }

  async function fetchUsersPage(
    subscribers: boolean,
    opts: Partial<AdminListQuery> = {},
  ) {
    const loadingKey = subscribers ? "subscribersLoading" : "usersLoading";
    let query: AdminListQuery = { ...defaultListQuery };

    update((state) => {
      const current = subscribers ? state.subscribersQuery : state.usersQuery;
      query = {
        page: opts.page ?? current.page,
        perPage: opts.perPage ?? current.perPage,
        search: opts.search ?? current.search,
      };
      const next = {
        ...state,
        error: null,
        [loadingKey]: true,
      } as AdminState;
      if (subscribers) {
        next.subscribersQuery = query;
      } else {
        next.usersQuery = query;
      }
      next.isLoading = anyLoading(next);
      return next;
    });

    const params = new URLSearchParams();
    params.set("page", String(query.page));
    params.set("per_page", String(query.perPage));
    if (query.search) params.set("search", query.search);
    if (subscribers) params.set("subscribers", "true");

    const { data, error } = await api.get<AdminUserList>(
      `/api/admin/users?${params.toString()}`,
    );

    update((state) => {
      const next: AdminState = {
        ...state,
        [loadingKey]: false,
        error: error ?? null,
      } as AdminState;
      if (!error && data) {
        if (subscribers) {
          next.subscribers = data.users || [];
          next.subscribersTotal = data.total;
          next.subscribersTotalPages = data.totalPages;
          next.subscribersQuery = {
            page: data.page,
            perPage: data.perPage,
            search: query.search,
          };
          next.subscribersLoaded = true;
        } else {
          next.users = data.users || [];
          next.usersTotal = data.total;
          next.usersTotalPages = data.totalPages;
          next.usersQuery = {
            page: data.page,
            perPage: data.perPage,
            search: query.search,
          };
          next.usersLoaded = true;
        }
      }
      next.isLoading = anyLoading(next);
      return next;
    });
  }

  return {
    subscribe,
    fetchUsers: async (opts: Partial<AdminListQuery> = {}) => {
      await fetchUsersPage(false, opts);
    },

    fetchSubscribers: async (opts: Partial<AdminListQuery> = {}) => {
      await fetchUsersPage(true, opts);
    },

    fetchContainers: async () => {
      update((state) => {
        const next = { ...state, containersLoading: true, error: null };
        next.isLoading = anyLoading(next);
        return next;
      });
      const { data, error } = await api.get<AdminContainer[]>(
        "/api/admin/containers",
      );

      update((state) => {
        const next: AdminState = {
          ...state,
          containersLoading: false,
          containersLoaded: !error,
          error: error ?? null,
        };
        if (!error) next.containers = data || [];
        next.isLoading = anyLoading(next);
        return next;
      });
    },

    fetchTerminals: async () => {
      update((state) => {
        const next = { ...state, terminalsLoading: true, error: null };
        next.isLoading = anyLoading(next);
        return next;
      });
      const { data, error } = await api.get<AdminTerminal[]>(
        "/api/admin/terminals",
      );

      update((state) => {
        const next: AdminState = {
          ...state,
          terminalsLoading: false,
          terminalsLoaded: !error,
          error: error ?? null,
        };
        if (!error) next.terminals = data || [];
        next.isLoading = anyLoading(next);
        return next;
      });
    },

    fetchAgents: async () => {
      update((state) => {
        const next = { ...state, agentsLoading: true, error: null };
        next.isLoading = anyLoading(next);
        return next;
      });
      const { data, error } = await api.get<AdminAgent[]>("/api/admin/agents");

      update((state) => {
        const next: AdminState = {
          ...state,
          agentsLoading: false,
          agentsLoaded: !error,
          error: error ?? null,
        };
        if (!error) next.agents = data || [];
        next.isLoading = anyLoading(next);
        return next;
      });
    },

    fetchStats: async (range = "30d") => {
      update((state) => {
        const next = { ...state, statsLoading: true, error: null };
        next.isLoading = anyLoading(next);
        return next;
      });
      const { data, error } = await api.get<AdminUsageStats>(
        `/api/admin/stats?range=${encodeURIComponent(range)}`,
      );

      update((state) => {
        const next: AdminState = {
          ...state,
          statsLoading: false,
          statsLoaded: !error,
          error: error ?? null,
        };
        if (!error) next.stats = data || null;
        next.isLoading = anyLoading(next);
        return next;
      });
    },

    deleteUser: async (userId: string) => {
      const { error } = await api.delete(`/api/admin/users/${userId}`);
      if (error) return { success: false, error };
      scheduleUsersRefetch();
      return { success: true };
    },

    deleteContainer: async (containerId: string) => {
      const { error } = await api.delete(
        `/api/admin/containers/${containerId}`,
      );
      if (error) return { success: false, error };
      // WS event will handle updating the store
      return { success: true };
    },

    startAdminEvents: () => {
      connectWebSocket();
    },

    stopAdminEvents: () => {
      disconnectWebSocket();
    },
  };
}

export const admin = createAdminStore();

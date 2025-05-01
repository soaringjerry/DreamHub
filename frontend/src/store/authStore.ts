import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { loginUser as loginUserApi, LoginCredentials, LoginResponse, RegisterPayload, registerUser as registerUserApi, RegisterResponse, SanitizedUser } from '../services/api'; // Import API functions and types

// --- State Interface ---
interface AuthState {
  isAuthenticated: boolean;
  user: SanitizedUser | null;
  token: string | null;
  isLoading: boolean; // For login/register loading state
  error: string | null;   // For login/register errors
}

// --- Actions Interface ---
interface AuthActions {
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
  register: (payload: RegisterPayload) => Promise<void>;
  // Internal action to set state after successful login/token load
  _setStateOnLogin: (token: string, user: SanitizedUser) => void;
  // Action to clear errors
  clearError: () => void;
}

// --- Initial State ---
const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  token: null,
  isLoading: false,
  error: null,
};

// --- Zustand Store Implementation with Persistence ---
export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      // --- Actions ---
      _setStateOnLogin: (token, user) => {
        set({ isAuthenticated: true, user, token, isLoading: false, error: null });
        // Note: Token is automatically persisted by the middleware
        console.log('Auth state updated on login:', user.username);
      },

      login: async (credentials) => {
        set({ isLoading: true, error: null });
        try {
          const response = await loginUserApi(credentials);
          get()._setStateOnLogin(response.token, response.user);
        } catch (error) {
          const errorMsg = error instanceof Error ? error.message : 'Login failed due to an unknown error.';
          set({ isLoading: false, error: errorMsg, isAuthenticated: false, user: null, token: null });
          // Also clear token from storage explicitly on login failure? Persist middleware might handle this.
          // localStorage.removeItem('auth-storage'); // Or whatever the key is
          console.error("Login failed in store action:", error);
        }
      },

      logout: () => {
        console.log('Logging out user:', get().user?.username);
        set({ ...initialState }); // Reset state to initial values
        // The persist middleware should handle clearing the token from localStorage
        // based on the state reset. If not, uncomment below:
        // localStorage.removeItem('auth-storage');
      },

      register: async (payload) => {
          set({ isLoading: true, error: null });
          try {
              await registerUserApi(payload);
              // Optionally automatically log in after successful registration
              // Or just clear loading/error and let user log in manually
              set({ isLoading: false, error: null });
              console.log('Registration successful for:', payload.username);
              // Consider redirecting to login or showing a success message
          } catch (error) {
              const errorMsg = error instanceof Error ? error.message : 'Registration failed due to an unknown error.';
              set({ isLoading: false, error: errorMsg });
              console.error("Registration failed in store action:", error);
          }
      },

      clearError: () => {
        set({ error: null });
      }

    }),
    {
      name: 'auth-storage', // Key for localStorage
      storage: createJSONStorage(() => localStorage),
      // Only persist token and user data, isAuthenticated can be derived
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        // Persist isAuthenticated as well for quicker initial check?
        isAuthenticated: state.isAuthenticated,
      }),
      // Optional: Rehydrate logic (runs when store loads from storage)
      // onRehydrateStorage: (state) => {
      //   console.log("Hydration finished");
      //   return (state, error) => {
      //     if (error) {
      //       console.error("An error happened during hydration", error);
      //     } else {
      //       // You could potentially validate the token here with an API call
      //       // If token is invalid/expired, call logout()
      //       if (state?.token) {
      //          console.log("Rehydrated with token for user:", state.user?.username);
      //          // Example: validateTokenApi(state.token).catch(() => state.logout());
      //       } else {
      //          console.log("Rehydrated without token.");
      //          // Ensure state is clean if no token
      //          state.isAuthenticated = false;
      //          state.user = null;
      //          state.token = null;
      //       }
      //     }
      //   };
      // }
    }
  )
);

// --- Selectors ---
export const useIsAuthenticated = () => useAuthStore((state) => state.isAuthenticated);
export const useCurrentUser = () => useAuthStore((state) => state.user);
export const useAuthToken = () => useAuthStore((state) => state.token);
export const useAuthLoading = () => useAuthStore((state) => state.isLoading);
export const useAuthError = () => useAuthStore((state) => state.error);

// --- Initialize store (attempt rehydration) ---
// Zustand automatically handles rehydration when using persist middleware.
// You might want to call a validation function here if needed after initial load.
console.log("Initial auth state from storage:", useAuthStore.getState());
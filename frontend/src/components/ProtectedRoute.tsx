import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useIsAuthenticated, useAuthToken } from '../store/authStore'; // Import auth state hook

interface ProtectedRouteProps {
  // You can add props here if needed, e.g., required roles/permissions
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = () => {
  const isAuthenticated = useIsAuthenticated();
  const token = useAuthToken(); // Get token to check persistence
  const location = useLocation();

  // Check both the state and the persisted token for robustness,
  // especially during initial load before rehydration might be fully complete.
  const isTrulyAuthenticated = isAuthenticated && !!token;

  if (!isTrulyAuthenticated) {
    // Redirect them to the /login page, but save the current location they were
    // trying to go to when they were redirected. This allows us to send them
    // along to that page after they login, which is a nicer user experience
    // than dropping them off on the home page.
    console.log('ProtectedRoute: User not authenticated, redirecting to login.');
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // If authenticated, render the child routes
  return <Outlet />;
};

export default ProtectedRoute;
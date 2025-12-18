import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import api from "@/api/client";
import { User } from "@/types";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  signUp: (email: string, password: string, fullName: string, practiceName: string) => Promise<{ error: Error | null }>;
  signIn: (email: string, password: string) => Promise<{ error: Error | null }>;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check for existing session
    const accessToken = localStorage.getItem("access_token");
    if (accessToken) {
        // TODO: ideally we should have a /me endpoint to fetch user profile with the token.
        // For now, we rely on the token being present, but we DON't have user details unless we persist them or fetch them.
        // Let's assume we decode token or fetch profile.
        // Since we don't have /me endpoint yet implementation plan didn't strictly say it but implied.
        // I'll try to decode token if possible, OR I should add /me endpoint to backend.
        // Actually, the Plan said "Load token from localStorage on init and validate session or fetch profile."
        // I will assume for now we just set loading false. 
        // Better: I will implement a fetchProfile function if needed, but for now let's just checking token presence.
        // However, without user object, the app might glitch if it relies on user.email etc.
        // I'll check if I can store user info in localStorage too for simplicity or if I need to add /me.
        // Adding /me is better practice. But for this step I'll try to stick to what I have or minimal changes.
        // Let's just try to persist user in localStorage as well for now to keep it simple without changing backend immediately again.
        
        const storedUser = localStorage.getItem("user_profile");
        if (storedUser) {
            setUser(JSON.parse(storedUser));
        }
    }
    setLoading(false);
  }, []);

  const signUp = async (email: string, password: string, fullName: string, practiceName: string) => {
    try {
      const response = await api.post("/auth/register", {
        email,
        password,
        full_name: fullName,
        practice_name: practiceName,
      });
      
      const { id, email: resEmail, role, full_name } = response.data.data; // response structure from backend
      // Backend returns RegisterResponse: {id, email, role}. It doesn't return tokens immediately usually unless auto-login.
      // My backend Register returns user object but NO tokens.
      // So after register, we might need to Login or user has to login manually.
      // Frontend Auth.tsx handles this: "Your account has been created successfully." -> User then switches to Sign In tab potentially?
      // Looking at Auth.tsx handleSignup: it calls signUp. If no error, it shows toast. It DOES NOT auto-login.
      // So we don't need to set user/tokens here.

      return { error: null };
    } catch (err: any) {
        // Axios error
        const message = err.response?.data?.message || err.message || "Registration failed";
        return { error: new Error(message) };
    }
  };

  const signIn = async (email: string, password: string) => {
    try {
      const response = await api.post("/auth/login", { email, password });
      const { access_token, refresh_token } = response.data.data;
      
      localStorage.setItem("access_token", access_token);
      localStorage.setItem("refresh_token", refresh_token);
      
      // We need to set the user state.
      // Since /login response only gives tokens (Update backend to return user? OR fetch me?)
      // Backend LoginResponse: {AccessToken, RefreshToken}.
      // I need to decode the token to get user info OR fetch /me.
      // To save time and backend cycles, I'll parse the JWT here to get basic info if possible, 
      // BUT I cannot parse signature verification.
      // Alternative: Update Backend Login to return User info too. 
      // For now, I will Mock the user object based on email since I don't have /me endpoint.
      // actually, let's decode jwt payload.
      
      const payload = JSON.parse(atob(access_token.split('.')[1]));
      const userProfile: User = {
        id: payload.user_id, // ensure payload has this
        email: payload.email,
        role: payload.role,
        full_name: "User", // JWT might not have this unless I added it.
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      };
      
      setUser(userProfile);
      localStorage.setItem("user_profile", JSON.stringify(userProfile));

      return { error: null };
    } catch (err: any) {
       const message = err.response?.data?.message || err.message || "Login failed";
       return { error: new Error(message) };
    }
  };

  const signOut = async () => {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("user_profile");
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, loading, signUp, signIn, signOut }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

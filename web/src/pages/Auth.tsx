import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Shield, Lock, Eye, EyeOff, ArrowLeft, Mail, CheckCircle2 } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useToast } from "@/hooks/use-toast";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import api from "@/api/client";

const loginSchema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(6, "Password must be at least 6 characters"),
});

const signupSchema = z.object({
  fullName: z.string().min(2, "Name must be at least 2 characters").max(100),
  practiceName: z.string().min(2, "Practice name must be at least 2 characters").max(100),
  email: z.string().email("Invalid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

const resetPasswordSchema = z.object({
  newPassword: z.string().min(8, "Password must be at least 8 characters"),
  confirmPassword: z.string(),
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

type AuthMode = "login" | "signup" | "forgot" | "reset" | "forgot-sent" | "reset-success";

const Auth = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { user, signIn, signUp, loading } = useAuth();
  const { toast } = useToast();
  const { t } = useTranslation('auth');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showSignupPassword, setShowSignupPassword] = useState(false);
  const [defaultTab, setDefaultTab] = useState<"login" | "signup">("login");

  // Determine initial mode from URL search params
  const urlMode = searchParams.get("mode");
  const resetToken = searchParams.get("token");
  const initialMode: AuthMode = urlMode === "reset" && resetToken ? "reset" : "login";
  const [mode, setMode] = useState<AuthMode>(initialMode);

  // Login form state
  const [loginEmail, setLoginEmail] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  
  // Signup form state
  const [signupFullName, setSignupFullName] = useState("");
  const [signupPracticeName, setSignupPracticeName] = useState("");
  const [signupEmail, setSignupEmail] = useState("");
  const [signupPassword, setSignupPassword] = useState("");
  const [signupConfirmPassword, setSignupConfirmPassword] = useState("");

  // Forgot password state
  const [forgotEmail, setForgotEmail] = useState("");

  // Reset password state
  const [resetNewPassword, setResetNewPassword] = useState("");
  const [resetConfirmPassword, setResetConfirmPassword] = useState("");

  useEffect(() => {
    if (user && !loading) {
      navigate("/dashboard");
    }
  }, [user, loading, navigate]);

  useEffect(() => {
    if (searchParams.get("signup") === "true") {
      setDefaultTab("signup");
    }
  }, [searchParams]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      loginSchema.parse({ email: loginEmail, password: loginPassword });
      
      const { error } = await signIn(loginEmail, loginPassword);
      
      if (error) {
        if (error.message.includes("Invalid login credentials")) {
          toast({
            title: t('login.failed'),
            description: t('login.invalidCredentials'),
            variant: "destructive",
          });
        } else {
          toast({
            title: t('login.failed'),
            description: error.message,
            variant: "destructive",
          });
        }
      } else {
        // Role-based redirect after successful login
        const stored = localStorage.getItem("user_profile");
        const profile = stored ? JSON.parse(stored) : null;
        if (profile?.system_role === "admin") {
          navigate("/admin");
        } else {
          navigate("/dashboard");
        }
      }
    } catch (err) {
      if (err instanceof z.ZodError) {
        toast({
          title: t('validation.validationError'),
          description: err.errors[0].message,
          variant: "destructive",
        });
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleSignup = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      signupSchema.parse({
        fullName: signupFullName,
        practiceName: signupPracticeName,
        email: signupEmail,
        password: signupPassword,
        confirmPassword: signupConfirmPassword,
      });
      
      const { error } = await signUp(signupEmail, signupPassword, signupFullName, signupPracticeName);
      
      if (error) {
        if (error.message.includes("already registered")) {
          toast({
            title: "Account Exists",
            description: "An account with this email already exists. Please sign in instead.",
            variant: "destructive",
          });
        } else {
          toast({
            title: "Signup Failed",
            description: error.message,
            variant: "destructive",
          });
        }
      } else {
        const planParam = searchParams.get("plan");
        const planMessage = planParam === "paid" 
          ? " You can upgrade to the paid plan anytime from your dashboard."
          : "";
        
        toast({
          title: "Welcome to Closaf!",
          description: `Your account has been created successfully. You're starting with our free tier.${planMessage}`,
        });
        navigate("/onboarding");
      }
    } catch (err) {
      if (err instanceof z.ZodError) {
        toast({
          title: "Validation Error",
          description: err.errors[0].message,
          variant: "destructive",
        });
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleForgotPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      z.string().email("Invalid email address").parse(forgotEmail);
      
      await api.post("/auth/forgot-password", { email: forgotEmail });
      setMode("forgot-sent");
    } catch (err) {
      if (err instanceof z.ZodError) {
        toast({
          title: "Validation Error",
          description: err.errors[0].message,
          variant: "destructive",
        });
      } else {
        // Still show success to prevent email enumeration
        setMode("forgot-sent");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleResetPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      resetPasswordSchema.parse({
        newPassword: resetNewPassword,
        confirmPassword: resetConfirmPassword,
      });

      await api.post("/auth/reset-password", {
        token: resetToken,
        new_password: resetNewPassword,
      });

      setMode("reset-success");
    } catch (err: unknown) {
      if (err instanceof z.ZodError) {
        toast({
          title: "Validation Error",
          description: err.errors[0].message,
          variant: "destructive",
        });
      } else {
        const message = (err as { response?: { data?: { message?: string } } })?.response?.data?.message || "Failed to reset password. The link may be expired.";
        toast({
          title: "Reset Failed",
          description: message,
          variant: "destructive",
        });
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-pulse">Loading...</div>
      </div>
    );
  }

  // ---- Forgot Password: email sent confirmation ----
  if (mode === "forgot-sent") {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-8">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
              <Mail className="h-7 w-7 text-primary" />
            </div>
            <CardTitle className="text-2xl">Check Your Email</CardTitle>
            <CardDescription>
              If an account exists for <strong>{forgotEmail}</strong>, we've sent a password reset link.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground text-center">
              The link will expire in 1 hour. Check your spam folder if you don't see it.
            </p>
            <Button
              variant="outline"
              className="w-full"
              onClick={() => { setMode("login"); setForgotEmail(""); }}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back to Sign In
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // ---- Reset Password: success confirmation ----
  if (mode === "reset-success") {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-8">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30">
              <CheckCircle2 className="h-7 w-7 text-green-600 dark:text-green-400" />
            </div>
            <CardTitle className="text-2xl">Password Reset</CardTitle>
            <CardDescription>
              Your password has been reset successfully. You can now sign in with your new password.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              className="w-full"
              onClick={() => { setMode("login"); navigate("/auth"); }}
            >
              Sign In
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // ---- Forgot Password: enter email ----
  if (mode === "forgot") {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-8">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
              <Lock className="h-7 w-7 text-primary" />
            </div>
            <CardTitle className="text-2xl">Forgot Password?</CardTitle>
            <CardDescription>
              Enter your email address and we'll send you a link to reset your password.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleForgotPassword} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="forgot-email">Email</Label>
                <Input
                  id="forgot-email"
                  type="email"
                  placeholder="dr.smith@practice.com"
                  value={forgotEmail}
                  onChange={(e) => setForgotEmail(e.target.value)}
                  required
                  autoFocus
                />
              </div>
              <Button type="submit" className="w-full" disabled={isSubmitting}>
                {isSubmitting ? "Sending..." : "Send Reset Link"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                className="w-full"
                onClick={() => setMode("login")}
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Sign In
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    );
  }

  // ---- Reset Password: set new password (from email link) ----
  if (mode === "reset") {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-8">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
              <Lock className="h-7 w-7 text-primary" />
            </div>
            <CardTitle className="text-2xl">Set New Password</CardTitle>
            <CardDescription>
              Enter your new password below.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleResetPassword} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="reset-password">New Password</Label>
                <div className="relative">
                  <Input
                    id="reset-password"
                    type={showPassword ? "text" : "password"}
                    placeholder="••••••••"
                    value={resetNewPassword}
                    onChange={(e) => setResetNewPassword(e.target.value)}
                    required
                    autoFocus
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="absolute right-2 top-1/2 -translate-y-1/2 h-8 w-8"
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </Button>
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="reset-confirm">Confirm Password</Label>
                <Input
                  id="reset-confirm"
                  type="password"
                  placeholder="••••••••"
                  value={resetConfirmPassword}
                  onChange={(e) => setResetConfirmPassword(e.target.value)}
                  required
                />
              </div>
              <p className="text-xs text-muted-foreground">
                Password must be at least 8 characters long.
              </p>
              <Button type="submit" className="w-full" disabled={isSubmitting}>
                {isSubmitting ? "Resetting..." : "Reset Password"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    );
  }

  // ---- Default: Login / Signup tabs ----
  return (
    <div className="min-h-screen flex bg-background">
      {/* Left side - Branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-primary/10 via-accent to-background flex-col justify-center items-center p-12">
        <div className="max-w-md text-center">
          <img src="/logo_text_vertical.svg" alt="Closaf" className="h-24 mx-auto mb-8" />
          <h1 className="text-4xl font-bold mb-4">Closaf</h1>
          <p className="text-lg text-muted-foreground mb-8">
            Secure, sovereign, and private practice management for mental health professionals.
          </p>
          <div className="space-y-4 text-left">
            <div className="flex items-center gap-3 text-muted-foreground">
              <Shield className="w-5 h-5 text-primary" />
              <span>End-to-end encrypted clinical notes</span>
            </div>
            <div className="flex items-center gap-3 text-muted-foreground">
              <Lock className="w-5 h-5 text-primary" />
              <span>HIPAA & GDPR compliant</span>
            </div>
          </div>
        </div>
      </div>

      {/* Right side - Auth forms */}
      <div className="flex-1 flex items-center justify-center p-8">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <img src="/logo_text_vertical.svg" alt="Closaf" className="h-16 mx-auto mb-4 lg:hidden" />
            <CardTitle className="text-2xl">{t('login.title')}</CardTitle>
            <CardDescription>
              {t('login.subtitle')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue={defaultTab} value={defaultTab} onValueChange={(v) => setDefaultTab(v as "login" | "signup")} className="w-full">
              <TabsList className="grid w-full grid-cols-2 mb-6">
                <TabsTrigger value="login">{t('tab.login')}</TabsTrigger>
                <TabsTrigger value="signup">{t('tab.register')}</TabsTrigger>
              </TabsList>

              <TabsContent value="login">
                <form onSubmit={handleLogin} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="login-email">{t('login.email')}</Label>
                    <Input
                      id="login-email"
                      type="email"
                      placeholder={t('login.emailPlaceholder')}
                      value={loginEmail}
                      onChange={(e) => setLoginEmail(e.target.value)}
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="login-password">{t('login.password')}</Label>
                    <div className="relative">
                      <Input
                        id="login-password"
                        type={showPassword ? "text" : "password"}
                        placeholder="••••••••"
                        value={loginPassword}
                        onChange={(e) => setLoginPassword(e.target.value)}
                        required
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-2 top-1/2 -translate-y-1/2 h-8 w-8"
                        onClick={() => setShowPassword(!showPassword)}
                      >
                        {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </Button>
                    </div>
                  </div>

                  {/* Forgot Password Link */}
                  <div className="flex justify-end">
                    <Button
                      type="button"
                      variant="link"
                      className="px-0 h-auto text-sm text-muted-foreground hover:text-primary"
                      onClick={() => setMode("forgot")}
                    >
                      {t('login.forgotPassword')}
                    </Button>
                  </div>

                  <Button type="submit" className="w-full" disabled={isSubmitting}>
                    {isSubmitting ? t('login.submitting') : t('login.submit')}
                  </Button>
                </form>
              </TabsContent>

              <TabsContent value="signup">
                {searchParams.get("plan") === "paid" && (
                  <div className="mb-4 p-3 bg-primary/10 border border-primary/20 rounded-md">
                    <p className="text-sm text-muted-foreground">
                      Start with our free tier and upgrade to paid plan anytime from your dashboard.
                    </p>
                  </div>
                )}
                <form onSubmit={handleSignup} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="signup-name">{t('register.fullName')}</Label>
                    <Input
                      id="signup-name"
                      type="text"
                      placeholder={t('register.fullNamePlaceholder')}
                      value={signupFullName}
                      onChange={(e) => setSignupFullName(e.target.value)}
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signup-practice">Practice Name</Label>
                    <Input
                      id="signup-practice"
                      type="text"
                      placeholder="Mindful Therapy Center"
                      value={signupPracticeName}
                      onChange={(e) => setSignupPracticeName(e.target.value)}
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signup-email">{t('register.email')}</Label>
                    <Input
                      id="signup-email"
                      type="email"
                      placeholder="dr.smith@practice.com"
                      value={signupEmail}
                      onChange={(e) => setSignupEmail(e.target.value)}
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signup-password">{t('register.password')}</Label>
                    <div className="relative">
                      <Input
                        id="signup-password"
                        type={showSignupPassword ? "text" : "password"}
                        placeholder="••••••••"
                        value={signupPassword}
                        onChange={(e) => setSignupPassword(e.target.value)}
                        required
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-2 top-1/2 -translate-y-1/2 h-8 w-8"
                        onClick={() => setShowSignupPassword(!showSignupPassword)}
                        aria-label={showSignupPassword ? "Hide password" : "Show password"}
                      >
                        {showSignupPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </Button>
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signup-confirm">{t('register.confirmPassword')}</Label>
                    <div className="relative">
                      <Input
                        id="signup-confirm"
                        type={showSignupPassword ? "text" : "password"}
                        placeholder="••••••••"
                        value={signupConfirmPassword}
                        onChange={(e) => setSignupConfirmPassword(e.target.value)}
                        required
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-2 top-1/2 -translate-y-1/2 h-8 w-8"
                        onClick={() => setShowSignupPassword(!showSignupPassword)}
                        aria-label={showSignupPassword ? "Hide password" : "Show password"}
                      >
                        {showSignupPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </Button>
                    </div>
                  </div>
                  <p className="text-xs text-muted-foreground text-center">
                    {t('register.trustLine')}
                  </p>
                  <Button type="submit" className="w-full" disabled={isSubmitting}>
                    {isSubmitting ? t('register.submitting') : t('register.submit')}
                  </Button>
                </form>
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default Auth;

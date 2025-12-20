import { useEffect, useState } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useToast } from "@/hooks/use-toast";
import { teamService, type TeamInvitation } from "@/services/teamService";
import { CheckCircle2, XCircle, Loader2, Mail, Eye, EyeOff } from "lucide-react";
import { format } from "date-fns";

const AcceptInvitation = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { toast } = useToast();
  const token = searchParams.get("token");

  const [invitation, setInvitation] = useState<TeamInvitation | null>(null);
  const [loading, setLoading] = useState(true);
  const [accepting, setAccepting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [signupData, setSignupData] = useState({
    fullName: "",
    password: "",
    confirmPassword: "",
  });

  useEffect(() => {
    if (!token) {
      setError("Invalid invitation link. No token provided.");
      setLoading(false);
      return;
    }

    loadInvitation();
  }, [token]);

  const loadInvitation = async () => {
    if (!token) return;

    try {
      const data = await teamService.getInvitationByToken(token);
      setInvitation(data);

      // Check if expired
      if (new Date(data.expires_at) < new Date()) {
        setError("This invitation has expired. Please ask for a new invitation.");
      } else if (data.status !== "pending") {
        setError(`This invitation has already been ${data.status}.`);
      }
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = error.response?.data?.error?.message || error.message || "Failed to load invitation";
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  const handleRegisterAndAccept = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token || !invitation) return;

    // Validate password match
    if (signupData.password !== signupData.confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    if (signupData.password.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }

    if (signupData.fullName.length < 2) {
      setError("Full name must be at least 2 characters");
      return;
    }

    setAccepting(true);
    setError(null);
    try {
      await teamService.registerAndAcceptInvitation({
        token,
        email: invitation.email,
        password: signupData.password,
        full_name: signupData.fullName,
      });
      
      toast({
        title: "Success",
        description: "Account created and invitation accepted! Please sign in.",
      });
      
      // Redirect to login after a short delay
      setTimeout(() => {
        navigate("/auth");
      }, 2000);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = error.response?.data?.error?.message || error.message || "Failed to create account";
      setError(message);
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setAccepting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-muted">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6">
            <div className="flex items-center justify-center gap-2">
              <Loader2 className="w-4 h-4 animate-spin" />
              <span>Loading invitation...</span>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (error && !invitation) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-muted">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-destructive">
              <XCircle className="w-5 h-5" />
              Invalid Invitation
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
            <Button onClick={() => navigate("/auth")} className="w-full">
              Go to Login
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!invitation) {
    return null;
  }

  const isExpired = new Date(invitation.expires_at) < new Date();
  const canAccept = invitation.status === "pending" && !isExpired;

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Mail className="w-5 h-5" />
            Team Invitation
          </CardTitle>
          <CardDescription>
            Create your account to join the organization
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {!canAccept ? (
            <Alert variant="destructive">
              <AlertDescription>
                {isExpired
                  ? "This invitation has expired. Please ask for a new invitation."
                  : `This invitation has already been ${invitation.status}.`}
              </AlertDescription>
            </Alert>
          ) : (
            <>
              <div className="space-y-2">
                <div className="text-sm">
                  <span className="font-medium">Email:</span> {invitation.email}
                </div>
                <div className="text-sm">
                  <span className="font-medium">Role:</span>{" "}
                  <span className="capitalize">{invitation.role.replace("_", " ")}</span>
                </div>
                <div className="text-sm">
                  <span className="font-medium">Expires:</span>{" "}
                  {format(new Date(invitation.expires_at), "MMM d, yyyy 'at' h:mm a")}
                </div>
              </div>

              <Alert>
                <AlertDescription>
                  This invitation is for new users only. If you already have an account, you cannot accept this invitation as you already belong to your own organization.
                </AlertDescription>
              </Alert>

              {canAccept && (
            <form onSubmit={handleRegisterAndAccept} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="fullName">Full Name</Label>
                <Input
                  id="fullName"
                  type="text"
                  placeholder="John Doe"
                  value={signupData.fullName}
                  onChange={(e) => setSignupData({ ...signupData, fullName: e.target.value })}
                  required
                  minLength={2}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={invitation.email}
                  disabled
                  className="bg-muted"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <div className="relative">
                  <Input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    placeholder="••••••••"
                    value={signupData.password}
                    onChange={(e) => setSignupData({ ...signupData, password: e.target.value })}
                    required
                    minLength={8}
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
                <Label htmlFor="confirmPassword">Confirm Password</Label>
                <Input
                  id="confirmPassword"
                  type={showPassword ? "text" : "password"}
                  placeholder="••••••••"
                  value={signupData.confirmPassword}
                  onChange={(e) => setSignupData({ ...signupData, confirmPassword: e.target.value })}
                  required
                  minLength={8}
                />
              </div>
              <Button type="submit" className="w-full" disabled={accepting}>
                {accepting ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Creating Account...
                  </>
                ) : (
                  <>
                    <CheckCircle2 className="w-4 h-4 mr-2" />
                    Create Account & Accept
                  </>
                )}
              </Button>
            </form>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default AcceptInvitation;


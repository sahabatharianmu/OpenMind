import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import OnboardingWizard from "@/components/onboarding/OnboardingWizard";
import GettingStartedModal from "@/components/onboarding/GettingStartedModal";
import { useState } from "react";

const Onboarding = () => {
  const { user, loading } = useAuth();
  const navigate = useNavigate();
  const [showGettingStarted, setShowGettingStarted] = useState(false);

  useEffect(() => {
    // Redirect to dashboard if not authenticated
    if (!loading && !user) {
      navigate("/auth");
    }
  }, [user, loading, navigate]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-pulse">Loading...</div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  const handleComplete = () => {
    // Mark onboarding as completed
    localStorage.setItem("onboarding_completed", "true");
    // Show getting started modal
    setShowGettingStarted(true);
  };

  const handleSkip = () => {
    // Mark onboarding as skipped (but not completed)
    localStorage.setItem("onboarding_skipped", "true");
    navigate("/dashboard");
  };

  const handleGettingStartedClose = () => {
    setShowGettingStarted(false);
    navigate("/dashboard");
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <div className="w-full max-w-2xl">
        <div className="text-center mb-6">
          <h1 className="text-2xl font-bold mb-2">Welcome to Closaf! 👋</h1>
          <p className="text-muted-foreground">
            Let's set up your practice. This will only take a few minutes.
          </p>
        </div>
        <OnboardingWizard onComplete={handleComplete} onSkip={handleSkip} />
      </div>
      {showGettingStarted && (
        <GettingStartedModal onClose={handleGettingStartedClose} />
      )}
    </div>
  );
};

export default Onboarding;


import { useState, useEffect } from "react";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import StatsCards from "@/components/dashboard/StatsCards";
import PatientList from "@/components/dashboard/PatientList";
import UpcomingAppointments from "@/components/dashboard/UpcomingAppointments";
import { useAuth } from "@/contexts/AuthContext";

const Dashboard = () => {
  const { user } = useAuth();
  
  const getGreeting = () => {
    const hour = new Date().getHours();
    if (hour < 12) return "Good morning";
    if (hour < 18) return "Good afternoon";
    return "Good evening";
  };

  // Use user's full_name from context, defaulting to "there"
  // Note: user.full_name might be "User" if not set in token/context correctly yet.
  const firstName = user?.full_name?.split(" ")[0] || "there";

  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-2xl lg:text-3xl font-bold">
            {getGreeting()}, {firstName}
          </h1>
          <p className="text-muted-foreground mt-1">
            Here's what's happening in your practice today.
          </p>
        </div>

        {/* Stats */}
        <StatsCards />

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-6 mt-6">
          <UpcomingAppointments />
          <PatientList />
        </div>
      </div>
    </DashboardLayout>
  );
};

export default Dashboard;

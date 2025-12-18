import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { Shield } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

const AuditLogs = () => {
  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl lg:text-3xl font-bold flex items-center gap-2">
              <Shield className="w-7 h-7" />
              Audit Logs
            </h1>
            <p className="text-muted-foreground mt-1">
              Track all data access and changes for compliance
            </p>
          </div>
        </div>

        <Card>
          <CardContent className="p-12 text-center">
            <Shield className="w-16 h-16 mx-auto mb-4 text-muted-foreground opacity-50" />
            <h2 className="text-xl font-semibold mb-2">Coming Soon</h2>
            <p className="text-muted-foreground max-w-md mx-auto">
              We are currently upgrading our audit logging system to provide better tracking and compliance features. This section will be available soon.
            </p>
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  );
};

export default AuditLogs;

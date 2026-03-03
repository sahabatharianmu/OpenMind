import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Users, CreditCard, Activity, Building2, Loader2, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import AdminLayout from "@/layouts/AdminLayout";
import { adminService, type AdminStats, type TenantListItem, type UserListItem } from "@/services/adminService";
import { useFormatters } from "@/hooks/useFormatters";

export default function AdminDashboard() {
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [recentTenants, setRecentTenants] = useState<TenantListItem[]>([]);
  const [recentUsers, setRecentUsers] = useState<UserListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const { t } = useTranslation('admin');
  const { t: tc } = useTranslation('common');
  const { formatCurrencyRaw, formatDate } = useFormatters();

  useEffect(() => {
    const fetchAll = async () => {
      try {
        const [statsData, tenantsData, usersData] = await Promise.all([
          adminService.getStats(),
          adminService.listTenants(1, 5),
          adminService.listUsers(1, 5),
        ]);
        setStats(statsData);
        setRecentTenants(tenantsData.tenants);
        setRecentUsers(usersData.users);
      } catch (error) {
        console.error("Failed to load admin data:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchAll();
  }, []);

  const tierBadge = (tier: string) => {
    const variants: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
      free: "outline",
      starter: "secondary",
      professional: "default",
      enterprise: "default",
    };
    return <Badge variant={variants[tier] || "outline"}>{tier || "free"}</Badge>;
  };

  return (
    <AdminLayout>
      <div className="space-y-6">
        <h2 className="text-3xl font-bold tracking-tight">{t('dashboard.title')}</h2>

        {/* Stats Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{t('dashboard.totalTenants')}</CardTitle>
              <Building2 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {loading ? (
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              ) : (
                <div className="text-2xl font-bold">{stats?.total_tenants ?? 0}</div>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{t('dashboard.totalUsers')}</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {loading ? (
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              ) : (
                <div className="text-2xl font-bold">{stats?.total_users ?? 0}</div>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{t('dashboard.activeSubscriptions')}</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {loading ? (
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              ) : (
                <div className="text-2xl font-bold">{stats?.active_subscriptions ?? 0}</div>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{t('dashboard.monthlyRevenue')}</CardTitle>
              <CreditCard className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {loading ? (
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              ) : (
                <div className="text-2xl font-bold">{formatCurrencyRaw(stats?.monthly_revenue ?? 0, 'IDR')}</div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Recent Tenants & Recent Users */}
        <div className="grid gap-6 lg:grid-cols-2">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>{t('dashboard.recentOrgs')}</CardTitle>
              <Button variant="ghost" size="sm" onClick={() => navigate("/admin/tenants")}>
                {tc('buttons.viewAll')} <ArrowRight className="ml-1 h-4 w-4" />
              </Button>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                </div>
              ) : recentTenants.length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-8">{t('dashboard.noOrgs')}</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('tenants.name')}</TableHead>
                      <TableHead>{t('tenants.plan')}</TableHead>
                      <TableHead className="text-right">{t('tenants.members')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {recentTenants.map((tenant) => (
                      <TableRow key={tenant.id}>
                        <TableCell className="font-medium">{tenant.name}</TableCell>
                        <TableCell>{tierBadge(tenant.subscription_tier)}</TableCell>
                        <TableCell className="text-right">{tenant.member_count}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>{t('dashboard.recentUsers')}</CardTitle>
              <Button variant="ghost" size="sm" onClick={() => navigate("/admin/tenants")}>
                {tc('buttons.viewAll')} <ArrowRight className="ml-1 h-4 w-4" />
              </Button>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                </div>
              ) : recentUsers.length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-8">{t('dashboard.noUsers')}</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('users.name')}</TableHead>
                      <TableHead>{t('users.email')}</TableHead>
                      <TableHead>{t('users.joined')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {recentUsers.map((u) => (
                      <TableRow key={u.id}>
                        <TableCell className="font-medium">{u.full_name}</TableCell>
                        <TableCell className="text-muted-foreground">{u.email}</TableCell>
                        <TableCell>{formatDate(u.created_at)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </AdminLayout>
  );
}

import { useState, useEffect, useCallback } from "react";
import AdminLayout from "@/layouts/AdminLayout";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Loader2, ChevronLeft, ChevronRight } from "lucide-react";
import { useTranslation } from "react-i18next";
import { adminService, type TenantListItem, type TenantListResponse } from "@/services/adminService";
import { useFormatters } from "@/hooks/useFormatters";

export default function Tenants() {
  const [data, setData] = useState<TenantListResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const { t } = useTranslation('admin');
  const { t: tc } = useTranslation('common');
  const { formatDate } = useFormatters();

  const fetchTenants = useCallback(async (p: number) => {
    try {
      setLoading(true);
      const res = await adminService.listTenants(p, 20);
      setData(res);
    } catch (error) {
      console.error("Failed to load tenants:", error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTenants(page);
  }, [fetchTenants, page]);

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
      <div className="mb-6">
        <h2 className="text-3xl font-bold tracking-tight">{t('tenants.title')}</h2>
        <p className="text-muted-foreground">{t('tenants.subtitle')}</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>
            {t('tenants.orgList')}
            {data && <span className="ml-2 text-sm font-normal text-muted-foreground">({t('tenants.total', { count: data.total })})</span>}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : !data || data.tenants.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              {t('tenants.noOrgs')}
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('tenants.name')}</TableHead>
                    <TableHead>{t('tenants.plan')}</TableHead>
                    <TableHead className="text-right">{t('tenants.members')}</TableHead>
                    <TableHead>{t('tenants.joined')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.tenants.map((tenant: TenantListItem) => (
                    <TableRow key={tenant.id}>
                      <TableCell className="font-medium">{tenant.name}</TableCell>
                      <TableCell>{tierBadge(tenant.subscription_tier)}</TableCell>
                      <TableCell className="text-right">{tenant.member_count}</TableCell>
                      <TableCell>{formatDate(tenant.created_at)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {/* Pagination */}
              {data.total_pages > 1 && (
                <div className="flex items-center justify-between mt-4">
                  <p className="text-sm text-muted-foreground">
                    {tc('pagination.page', { current: data.page, total: data.total_pages })}
                  </p>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={page <= 1}
                      onClick={() => setPage((p) => p - 1)}
                    >
                      <ChevronLeft className="h-4 w-4" /> {tc('pagination.previous')}
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={page >= data.total_pages}
                      onClick={() => setPage((p) => p + 1)}
                    >
                      {tc('pagination.next')} <ChevronRight className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </AdminLayout>
  );
}

import { useMemo } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { TrendingUp } from "lucide-react";
import { format, startOfMonth, subMonths, parseISO, isWithinInterval } from "date-fns";

interface Invoice {
  id: string;
  amount_cents: number;
  status: string;
  paid_at?: string | null;
  created_at: string;
}

interface RevenueChartProps {
  invoices: Invoice[];
}

export const RevenueChart = ({ invoices }: RevenueChartProps) => {
  const chartData = useMemo(() => {
    const months: { month: string; start: Date; end: Date }[] = [];
    const now = new Date();

    // Get last 6 months
    for (let i = 5; i >= 0; i--) {
      const monthStart = startOfMonth(subMonths(now, i));
      const monthEnd = new Date(monthStart);
      monthEnd.setMonth(monthEnd.getMonth() + 1);
      monthEnd.setDate(0); // Last day of month

      months.push({
        month: format(monthStart, "MMM"),
        start: monthStart,
        end: monthEnd,
      });
    }

    return months.map(({ month, start, end }) => {
      const monthInvoices = invoices.filter((inv) => {
        if (inv.status !== "paid" || !inv.paid_at) return false;
        const paidDate = parseISO(inv.paid_at);
        return isWithinInterval(paidDate, { start, end });
      });

      const revenue = monthInvoices.reduce((sum, inv) => sum + inv.amount_cents, 0);

      return {
        month,
        revenue: revenue / 100,
      };
    });
  }, [invoices]);

  const totalRevenue = chartData.reduce((sum, d) => sum + d.revenue, 0);

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-medium flex items-center gap-2">
            <TrendingUp className="w-4 h-4 text-primary" />
            Revenue (Last 6 Months)
          </CardTitle>
          <span className="text-lg font-bold">
            {new Intl.NumberFormat("en-US", {
              style: "currency",
              currency: "USD",
            }).format(totalRevenue)}
          </span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="h-[200px]">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData}>
              <defs>
                <linearGradient id="revenueGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis
                dataKey="month"
                tick={{ fontSize: 12 }}
                tickLine={false}
                axisLine={false}
                className="fill-muted-foreground"
              />
              <YAxis
                tick={{ fontSize: 12 }}
                tickLine={false}
                axisLine={false}
                tickFormatter={(value) => `$${value}`}
                className="fill-muted-foreground"
              />
              <Tooltip
                content={({ active, payload }) => {
                  if (active && payload && payload.length) {
                    return (
                      <div className="rounded-lg border bg-background p-2 shadow-sm">
                        <div className="text-sm font-medium">
                          {new Intl.NumberFormat("en-US", {
                            style: "currency",
                            currency: "USD",
                          }).format(payload[0].value as number)}
                        </div>
                      </div>
                    );
                  }
                  return null;
                }}
              />
              <Area
                type="monotone"
                dataKey="revenue"
                stroke="hsl(var(--primary))"
                strokeWidth={2}
                fill="url(#revenueGradient)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
};

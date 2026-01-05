import AdminLayout from "@/layouts/AdminLayout";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function Tenants() {
  return (
    <AdminLayout>
       <div className="mb-6">
          <h2 className="text-3xl font-bold tracking-tight">Tenants</h2>
          <p className="text-muted-foreground">Manage registered organizations.</p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Organization List</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Plan</TableHead>
                  <TableHead>Members</TableHead>
                  <TableHead>Joined Date</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {/* Placeholders */}
                <TableRow>
                  <TableCell className="font-medium">Acme Corp</TableCell>
                  <TableCell><Badge>Active</Badge></TableCell>
                  <TableCell>Pro Plan</TableCell>
                  <TableCell>12</TableCell>
                  <TableCell>Oct 24, 2024</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell className="font-medium">Globex Inc</TableCell>
                  <TableCell><Badge variant="secondary">Trial</Badge></TableCell>
                  <TableCell>Basic Plan</TableCell>
                  <TableCell>4</TableCell>
                  <TableCell>Nov 11, 2024</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
        </Card>
    </AdminLayout>
  );
}

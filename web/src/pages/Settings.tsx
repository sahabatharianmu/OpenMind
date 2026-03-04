import { useState, useEffect } from "react";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { User, Building2, Shield, Database, LogOut } from "lucide-react";
import { useAuth } from "@/contexts/AuthContext";
import { useToast } from "@/hooks/use-toast";
import { userService } from "@/services/userService";
import { organizationService } from "@/services/organizationService";
import { exportService } from "@/services/exportService";
import { importService, type ImportPreviewResponse, type ImportExecuteResponse } from "@/services/importService";
import type { UserProfile } from "@/services/userService";
import type { Organization } from "@/services/organizationService";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Upload, FileText, AlertCircle, CheckCircle2, XCircle, Download } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const Settings = () => {
  const { user, signOut } = useAuth();
  const { toast } = useToast();
  
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [fullName, setFullName] = useState("");
  const [loading, setLoading] = useState(false);
  
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [orgName, setOrgName] = useState("");
  const [taxId, setTaxId] = useState("");
  const [npi, setNpi] = useState("");
  const [address, setAddress] = useState("");
  const [currency, setCurrency] = useState("USD");
  const [locale, setLocale] = useState("en-US");
  const [orgLoading, setOrgLoading] = useState(false);
  
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordLoading, setPasswordLoading] = useState(false);

  // Import state
  const [importType, setImportType] = useState<"patients" | "appointments" | "notes">("patients");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewData, setPreviewData] = useState<ImportPreviewResponse | null>(null);
  const [importResult, setImportResult] = useState<ImportExecuteResponse | null>(null);
  const [importLoading, setImportLoading] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);

  useEffect(() => {
    loadProfile();
    loadOrganization();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const loadProfile = async () => {
    try {
      const data = await userService.getProfile();
      setProfile(data);
      setFullName(data.full_name);
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load profile",
        variant: "destructive",
      });
    }
  };

  const loadOrganization = async () => {
    try {
      const data = await organizationService.getMyOrganization();
      setOrganization(data);
      setOrgName(data.name);
      setTaxId(data.tax_id || "");
      setNpi(data.npi || "");
      setAddress(data.address || "");
      setCurrency(data.currency || "USD");
      setLocale(data.locale || "en-US");
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load organization",
        variant: "destructive",
      });
    }
  };

  const handleUpdateProfile = async () => {
    if (!fullName.trim()) {
      toast({
        title: "Error",
        description: "Name cannot be empty",
        variant: "destructive",
      });
      return;
    }

    setLoading(true);
    try {
      const updated = await userService.updateProfile({ full_name: fullName });
      setProfile(updated);
      toast({
        title: "Success",
        description: "Profile updated successfully",
      });
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } }; message?: string };
      const message = err.response?.data?.message || err.message || "Failed to update profile";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handlePreviewImport = async () => {
    if (!selectedFile) {
      toast({
        title: "Error",
        description: "Please select a file first",
        variant: "destructive",
      });
      return;
    }

    setPreviewLoading(true);
    try {
      const fileData = await fileToBase64(selectedFile);
      const preview = await importService.previewImport({
        type: importType,
        file_data: fileData,
        file_name: selectedFile.name,
      });
      setPreviewData(preview);
    } catch (error: unknown) {
      console.error("Preview error:", error);
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = err.response?.data?.error?.message || err.message || "Failed to preview import";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setPreviewLoading(false);
    }
  };

  const handleExecuteImport = async () => {
    if (!selectedFile || !previewData) {
      toast({
        title: "Error",
        description: "Please preview the import first",
        variant: "destructive",
      });
      return;
    }

    setImportLoading(true);
    try {
      const fileData = await fileToBase64(selectedFile);
      const result = await importService.executeImport({
        type: importType,
        file_data: fileData,
        file_name: selectedFile.name,
      });
      setImportResult(result);
      toast({
        title: "Success",
        description: `Imported ${result.success_count} of ${result.total_rows} records`,
      });
      // Reset for next import
      setSelectedFile(null);
      setPreviewData(null);
    } catch (error: unknown) {
      console.error("Import error:", error);
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = err.response?.data?.error?.message || err.message || "Failed to execute import";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setImportLoading(false);
    }
  };

  const fileToBase64 = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result as string;
        // Remove data URL prefix if present
        const base64 = result.includes(",") ? result.split(",")[1] : result;
        resolve(base64);
      };
      reader.onerror = reject;
      reader.readAsDataURL(file);
    });
  };

  const handleChangePassword = async () => {
    if (!oldPassword || !newPassword || !confirmPassword) {
      toast({
        title: "Error",
        description: "All fields are required",
        variant: "destructive",
      });
      return;
    }

    if (newPassword !== confirmPassword) {
      toast({
        title: "Error",
        description: "New passwords do not match",
        variant: "destructive",
      });
      return;
    }

    if (newPassword.length < 8) {
      toast({
        title: "Error",
        description: "Password must be at least 8 characters",
        variant: "destructive",
      });
      return;
    }

    setPasswordLoading(true);
    try {
      await userService.changePassword({
        old_password: oldPassword,
        new_password: newPassword,
      });
      
      setOldPassword("");
      setNewPassword("");
      setConfirmPassword("");
      
      toast({
        title: "Success",
        description: "Password changed successfully",
      });
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } }; message?: string };
      const message = err.response?.data?.message || err.message || "Failed to change password";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setPasswordLoading(false);
    }
  };

  const handleUpdateOrganization = async () => {
    if (!orgName.trim()) {
      toast({
        title: "Error",
        description: "Organization name cannot be empty",
        variant: "destructive",
      });
      return;
    }

    setOrgLoading(true);
    try {
      const updated = await organizationService.updateOrganization({ 
        name: orgName,
        tax_id: taxId,
        npi: npi,
        address: address,
        currency: currency,
        locale: locale
      });
      setOrganization(updated);
      toast({
        title: "Success",
        description: "Organization updated successfully",
      });
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } }; message?: string };
      const message = err.response?.data?.message || err.message || "Failed to update organization";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setOrgLoading(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8 max-w-4xl">
        <div className="mb-6">
          <h1 className="text-2xl lg:text-3xl font-bold">Settings</h1>
          <p className="text-muted-foreground mt-1">
            Manage your account and practice settings
          </p>
        </div>

        <Tabs defaultValue="profile" className="space-y-6">
          <TabsList>
            <TabsTrigger value="profile" className="gap-2">
              <User className="w-4 h-4" />
              Profile
            </TabsTrigger>
            <TabsTrigger value="practice" className="gap-2">
              <Building2 className="w-4 h-4" />
              Practice
            </TabsTrigger>
            <TabsTrigger value="security" className="gap-2">
              <Shield className="w-4 h-4" />
              Security
            </TabsTrigger>
            <TabsTrigger value="data" className="gap-2">
              <Database className="w-4 h-4" />
              Data
            </TabsTrigger>
          </TabsList>

          <TabsContent value="profile">
            <Card>
              <CardHeader>
                <CardTitle>Profile Information</CardTitle>
                <CardDescription>
                  Update your personal information
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="space-y-2">
                  <Label htmlFor="fullName">Full Name</Label>
                  <Input
                    id="fullName"
                    value={fullName}
                    onChange={(e) => setFullName(e.target.value)}
                    placeholder="Enter your full name"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    value={profile?.email || user?.email || ""}
                    disabled
                    className="bg-muted"
                  />
                  <p className="text-xs text-muted-foreground">
                    Email cannot be changed
                  </p>
                </div>
                <Button onClick={handleUpdateProfile} disabled={loading}>
                  {loading ? "Saving..." : "Save Changes"}
                </Button>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="practice">
            <Card>
              <CardHeader>
                <CardTitle>Practice Information</CardTitle>
                <CardDescription>
                  Manage your clinical and billing details. These are used for PDF Superbills.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="orgName">Organization / Practice Name</Label>
                    <Input
                      id="orgName"
                      value={orgName}
                      onChange={(e) => setOrgName(e.target.value)}
                      placeholder="Enter organization name"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Type</Label>
                    <Input
                      value={organization?.type || "-"}
                      disabled
                      className="bg-muted"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="taxId">Tax ID / EIN</Label>
                    <Input
                      id="taxId"
                      value={taxId}
                      onChange={(e) => setTaxId(e.target.value)}
                      placeholder="Enter Tax ID"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="npi">National Provider Identifier (NPI)</Label>
                    <Input
                      id="npi"
                      value={npi}
                      onChange={(e) => setNpi(e.target.value)}
                      placeholder="Enter NPI"
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="address">Practice Address</Label>
                  <Input
                    id="address"
                    value={address}
                    onChange={(e) => setAddress(e.target.value)}
                    placeholder="Street, City, State, ZIP"
                  />
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="currency">Currency</Label>
                    <Select value={currency} onValueChange={setCurrency}>
                      <SelectTrigger id="currency">
                        <SelectValue placeholder="Select currency" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="USD">USD - US Dollar</SelectItem>
                        <SelectItem value="EUR">EUR - Euro</SelectItem>
                        <SelectItem value="GBP">GBP - British Pound</SelectItem>
                        <SelectItem value="IDR">IDR - Indonesian Rupiah</SelectItem>
                        <SelectItem value="CAD">CAD - Canadian Dollar</SelectItem>
                        <SelectItem value="AUD">AUD - Australian Dollar</SelectItem>
                        <SelectItem value="JPY">JPY - Japanese Yen</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="locale">Display Locale (Date/Number Format)</Label>
                    <Select value={locale} onValueChange={setLocale}>
                      <SelectTrigger id="locale">
                        <SelectValue placeholder="Select locale" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="en-US">English (US)</SelectItem>
                        <SelectItem value="en-GB">English (UK)</SelectItem>
                        <SelectItem value="id-ID">Indonesian (Indonesia)</SelectItem>
                        <SelectItem value="de-DE">German (Germany)</SelectItem>
                        <SelectItem value="fr-FR">French (France)</SelectItem>
                        <SelectItem value="ja-JP">Japanese (Japan)</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <Separator />
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label>Member Count</Label>
                    <Input
                      value={organization?.member_count?.toString() || "0"}
                      disabled
                      className="bg-muted"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Created</Label>
                    <Input
                      value={organization?.created_at ? new Date(organization.created_at).toLocaleDateString() : "-"}
                      disabled
                      className="bg-muted"
                    />
                  </div>
                </div>
                <Button onClick={handleUpdateOrganization} disabled={orgLoading}>
                  {orgLoading ? "Saving..." : "Save Changes"}
                </Button>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="security">
            <Card>
              <CardHeader>
                <CardTitle>Security Settings</CardTitle>
                <CardDescription>
                  Manage your account security.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="space-y-4">
                  <div>
                    <h4 className="font-medium mb-3">Change Password</h4>
                    <div className="space-y-3">
                      <div className="space-y-2">
                        <Label htmlFor="oldPassword">Current Password</Label>
                        <Input
                          id="oldPassword"
                          type="password"
                          value={oldPassword}
                          onChange={(e) => setOldPassword(e.target.value)}
                          placeholder="Enter current password"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="newPassword">New Password</Label>
                        <Input
                          id="newPassword"
                          type="password"
                          value={newPassword}
                          onChange={(e) => setNewPassword(e.target.value)}
                          placeholder="Enter new password (min 8 characters)"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="confirmPassword">Confirm New Password</Label>
                        <Input
                          id="confirmPassword"
                          type="password"
                          value={confirmPassword}
                          onChange={(e) => setConfirmPassword(e.target.value)}
                          placeholder="Confirm new password"
                        />
                      </div>
                      <Button onClick={handleChangePassword} disabled={passwordLoading}>
                        {passwordLoading ? "Changing..." : "Change Password"}
                      </Button>
                    </div>
                  </div>
                  <Separator />
                  <div>
                    <h4 className="font-medium mb-1">Sign Out</h4>
                     <p className="text-sm text-muted-foreground mb-3">
                      Sign out of your session on this device.
                    </p>
                    <Button variant="destructive" onClick={signOut} className="gap-2">
                      <LogOut className="w-4 h-4" />
                      Sign Out
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="data">
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Practice Sovereignty</CardTitle>
                  <CardDescription>
                    Your data belongs to you. Export your entire practice history for backup or migration.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div>
                    <h4 className="font-medium mb-1">Clinic Takeout (Full Export)</h4>
                    <p className="text-sm text-muted-foreground mb-3">
                      Download a comprehensive ZIP archive containing all patients, appointments, clinical notes (including addendums), invoices, and an immutable audit log of all system access.
                    </p>
                    <Button 
                      onClick={async () => {
                        try {
                          const blob = await exportService.exportAllData();
                          
                          const url = window.URL.createObjectURL(new Blob([blob], { type: "application/zip" }));
                          const a = document.createElement("a");
                          a.href = url;
                          a.download = `openmind-export-${new Date().toISOString().split("T")[0]}.zip`;
                          document.body.appendChild(a);
                          a.click();
                          window.URL.revokeObjectURL(url);
                          document.body.removeChild(a);
                          
                          toast({
                            title: "Success",
                            description: "Data exported successfully",
                          });
                        } catch (error: unknown) {
                          console.error("Export error:", error);
                          const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
                          const message = err.response?.data?.error?.message || err.message || "Failed to export data";
                          toast({
                            title: "Error",
                            description: message,
                            variant: "destructive",
                          });
                        }
                      }}
                      className="gap-2"
                    >
                      <Database className="w-4 h-4" />
                      Download All Data
                    </Button>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Data Import</CardTitle>
                  <CardDescription>
                    Import patients, appointments, or clinical notes using our templates. Download a template, fill it with your data, then upload it here.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="importType">What would you like to import?</Label>
                      <Select value={importType} onValueChange={(v) => {
                        setImportType(v as "patients" | "notes");
                        setSelectedFile(null);
                        setPreviewData(null);
                        setImportResult(null);
                      }}>
                        <SelectTrigger id="importType">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="patients">Patients (CSV/XLSX)</SelectItem>
                          <SelectItem value="notes">Clinical Notes (CSV/XLSX)</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="border rounded-lg p-4 bg-muted/50 space-y-3">
                      <div className="flex items-center justify-between">
                        <div>
                          <h4 className="font-medium text-sm">Step 1: Download Template</h4>
                          <p className="text-xs text-muted-foreground mt-1">
                            Get the CSV or XLSX template file with the correct format
                          </p>
                        </div>
                        <div className="flex gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={async () => {
                              try {
                                const blob = await importService.downloadTemplate(importType, "csv");
                                const url = window.URL.createObjectURL(blob);
                                const a = document.createElement("a");
                                a.href = url;
                                a.download = importType === "patients" 
                                  ? "patients-import-template.csv"
                                  : "clinical-notes-import-template.csv";
                                document.body.appendChild(a);
                                a.click();
                                window.URL.revokeObjectURL(url);
                                document.body.removeChild(a);
                                toast({
                                  title: "Template Downloaded",
                                  description: "Fill in the template with your data, then upload it below.",
                                });
                              } catch (error) {
                                toast({
                                  title: "Error",
                                  description: "Failed to download template",
                                  variant: "destructive",
                                });
                              }
                            }}
                            className="gap-2"
                          >
                            <Download className="w-4 h-4" />
                            CSV
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={async () => {
                              try {
                                const blob = await importService.downloadTemplate(importType, "xlsx");
                                const url = window.URL.createObjectURL(blob);
                                const a = document.createElement("a");
                                a.href = url;
                                a.download = importType === "patients" 
                                  ? "patients-import-template.xlsx"
                                  : "clinical-notes-import-template.xlsx";
                                document.body.appendChild(a);
                                a.click();
                                window.URL.revokeObjectURL(url);
                                document.body.removeChild(a);
                                toast({
                                  title: "Template Downloaded",
                                  description: "Fill in the template with your data, then upload it below.",
                                });
                              } catch (error) {
                                toast({
                                  title: "Error",
                                  description: "Failed to download template",
                                  variant: "destructive",
                                });
                              }
                            }}
                            className="gap-2"
                          >
                            <Download className="w-4 h-4" />
                            XLSX
                          </Button>
                        </div>
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="importFile">Step 2: Upload Your File</Label>
                      <div className="flex items-center gap-4">
                        <Input
                          id="importFile"
                          type="file"
                          accept={importType === "patients" ? ".csv,.xlsx" : ".csv,.xlsx"}
                          onChange={(e) => {
                            const file = e.target.files?.[0];
                            if (file) {
                              setSelectedFile(file);
                              setPreviewData(null);
                              setImportResult(null);
                            }
                          }}
                          className="flex-1"
                        />
                        {selectedFile && (
                          <span className="text-sm text-muted-foreground">
                            {selectedFile.name}
                          </span>
                        )}
                      </div>
                    </div>

                    {selectedFile && (
                      <div className="flex gap-2">
                        <Button
                          onClick={handlePreviewImport}
                          disabled={previewLoading}
                          variant="outline"
                          className="gap-2"
                        >
                          <FileText className="w-4 h-4" />
                          {previewLoading ? "Previewing..." : "Step 3: Preview"}
                        </Button>
                        {previewData && (
                          <Button
                            onClick={handleExecuteImport}
                            disabled={importLoading}
                            className="gap-2"
                          >
                            <Upload className="w-4 h-4" />
                            {importLoading ? "Importing..." : "Step 4: Import"}
                          </Button>
                        )}
                      </div>
                    )}

                    {previewData && (
                      <div className="space-y-4">
                        <Alert>
                          <AlertCircle className="h-4 w-4" />
                          <AlertDescription>
                            <div className="space-y-1">
                              <div>Total Rows: {previewData.total_rows}</div>
                              <div className="flex items-center gap-2">
                                <CheckCircle2 className="w-4 h-4 text-green-600" />
                                Valid: {previewData.valid_rows}
                              </div>
                              {previewData.invalid_rows > 0 && (
                                <div className="flex items-center gap-2">
                                  <XCircle className="w-4 h-4 text-red-600" />
                                  Invalid: {previewData.invalid_rows}
                                </div>
                              )}
                            </div>
                          </AlertDescription>
                        </Alert>

                        {previewData.errors && previewData.errors.length > 0 && (
                          <div className="space-y-2">
                            <h4 className="font-medium text-sm">Errors</h4>
                            <div className="border rounded-md max-h-48 overflow-auto">
                              <Table>
                                <TableHeader>
                                  <TableRow>
                                    <TableHead>Row</TableHead>
                                    <TableHead>Field</TableHead>
                                    <TableHead>Message</TableHead>
                                  </TableRow>
                                </TableHeader>
                                <TableBody>
                                  {previewData.errors.slice(0, 10).map((error, idx) => (
                                    <TableRow key={idx}>
                                      <TableCell>{error.row}</TableCell>
                                      <TableCell>{error.field || "-"}</TableCell>
                                      <TableCell className="text-sm">{error.message}</TableCell>
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </div>
                            {previewData.errors.length > 10 && (
                              <p className="text-xs text-muted-foreground">
                                Showing first 10 errors. Total: {previewData.errors.length}
                              </p>
                            )}
                          </div>
                        )}

                        {previewData.warnings && previewData.warnings.length > 0 && (
                          <div className="space-y-2">
                            <h4 className="font-medium text-sm">Warnings</h4>
                            <div className="border rounded-md max-h-48 overflow-auto">
                              <Table>
                                <TableHeader>
                                  <TableRow>
                                    <TableHead>Row</TableHead>
                                    <TableHead>Field</TableHead>
                                    <TableHead>Message</TableHead>
                                  </TableRow>
                                </TableHeader>
                                <TableBody>
                                  {previewData.warnings.slice(0, 10).map((warning, idx) => (
                                    <TableRow key={idx}>
                                      <TableCell>{warning.row}</TableCell>
                                      <TableCell>{warning.field || "-"}</TableCell>
                                      <TableCell className="text-sm">{warning.message}</TableCell>
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </div>
                          </div>
                        )}

                        {previewData.preview && previewData.preview.length > 0 && (
                          <div className="space-y-2">
                            <h4 className="font-medium text-sm">Preview (First 10 rows)</h4>
                            <div className="border rounded-md max-h-64 overflow-auto">
                              <Table>
                                <TableHeader>
                                  <TableRow>
                                    {Object.keys(previewData.preview[0]).map((key) => (
                                      <TableHead key={key}>{key}</TableHead>
                                    ))}
                                  </TableRow>
                                </TableHeader>
                                <TableBody>
                                  {previewData.preview.map((row, idx) => (
                                    <TableRow key={idx}>
                                      {Object.values(row).map((val, cellIdx) => (
                                        <TableCell key={cellIdx} className="text-sm">
                                          {String(val || "-")}
                                        </TableCell>
                                      ))}
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </div>
                          </div>
                        )}
                      </div>
                    )}

                    {importResult && (
                      <Alert className={importResult.error_count === 0 ? "border-green-500" : ""}>
                        {importResult.error_count === 0 ? (
                          <CheckCircle2 className="h-4 w-4 text-green-600" />
                        ) : (
                          <AlertCircle className="h-4 w-4" />
                        )}
                        <AlertDescription>
                          <div className="space-y-1">
                            <div className="font-medium">
                              {importResult.error_count === 0 ? "Import Successful!" : "Import Completed with Errors"}
                            </div>
                            <div>Total: {importResult.total_rows}</div>
                            <div className="flex items-center gap-2">
                              <CheckCircle2 className="w-4 h-4 text-green-600" />
                              Imported: {importResult.success_count}
                            </div>
                            {importResult.error_count > 0 && (
                              <div className="flex items-center gap-2">
                                <XCircle className="w-4 h-4 text-red-600" />
                                Errors: {importResult.error_count}
                              </div>
                            )}
                            {importResult.errors && importResult.errors.length > 0 && (
                              <div className="mt-2 border rounded-md max-h-32 overflow-auto">
                                <Table>
                                  <TableHeader>
                                    <TableRow>
                                      <TableHead>Row</TableHead>
                                      <TableHead>Field</TableHead>
                                      <TableHead>Message</TableHead>
                                    </TableRow>
                                  </TableHeader>
                                  <TableBody>
                                    {importResult.errors.slice(0, 5).map((error, idx) => (
                                      <TableRow key={idx}>
                                        <TableCell>{error.row}</TableCell>
                                        <TableCell>{error.field || "-"}</TableCell>
                                        <TableCell className="text-sm">{error.message}</TableCell>
                                      </TableRow>
                                    ))}
                                  </TableBody>
                                </Table>
                              </div>
                            )}
                          </div>
                        </AlertDescription>
                      </Alert>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </DashboardLayout>
  );
};

export default Settings;

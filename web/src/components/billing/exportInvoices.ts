interface Invoice {
  id: string;
  amount_cents: number;
  status: string;
  due_date?: string | null;
  paid_at?: string | null;
  notes?: string | null;
  created_at: string;
  payment_method?: string | null;
  patients: {
    first_name: string;
    last_name: string;
  };
}

export const exportInvoicesToCSV = (invoices: Invoice[]) => {
  const headers = [
    "Invoice ID",
    "Patient Name",
    "Amount",
    "Status",
    "Due Date",
    "Paid Date",
    "Payment Method",
    "Created Date",
    "Notes",
  ];

  const rows = invoices.map((inv) => [
    inv.id,
    `${inv.patients.first_name} ${inv.patients.last_name}`,
    (inv.amount_cents / 100).toFixed(2),
    inv.status,
    inv.due_date || "",
    inv.paid_at ? new Date(inv.paid_at).toLocaleDateString() : "",
    inv.payment_method || "",
    new Date(inv.created_at).toLocaleDateString(),
    inv.notes || "",
  ]);

  const csvContent = [
    headers.join(","),
    ...rows.map((row) =>
      row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(",")
    ),
  ].join("\n");

  const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `invoices-${new Date().toISOString().split("T")[0]}.csv`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
};

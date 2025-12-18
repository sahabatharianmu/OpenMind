import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import patientService from "@/services/patientService";
import appointmentService from "@/services/appointmentService";
import clinicalNoteService from "@/services/clinicalNoteService";
import invoiceService from "@/services/invoiceService";
import { auditLogService, type AuditLogFilters } from "@/services/auditLogService";
import { CreatePatientRequest } from "@/services/patientService";

export const usePatients = () => {
  return useQuery({
    queryKey: ["patients"],
    queryFn: () => patientService.list(),
    staleTime: 60 * 1000, // 1 minute
  });
};

export const useAppointments = () => {
  return useQuery({
    queryKey: ["appointments"],
    queryFn: () => appointmentService.list(),
    staleTime: 60 * 1000,
  });
};

export const useClinicalNotes = () => {
  return useQuery({
    queryKey: ["clinical-notes"],
    queryFn: () => clinicalNoteService.list(),
    staleTime: 60 * 1000,
  });
};

export const useInvoices = () => {
  return useQuery({
    queryKey: ["invoices"],
    queryFn: () => invoiceService.list(),
    staleTime: 60 * 1000,
  });
};

export const useAuditLogs = (page: number = 1, pageSize: number = 20, filters?: AuditLogFilters) => {
  return useQuery({
    queryKey: ["audit-logs", page, pageSize, filters],
    queryFn: () => auditLogService.list(page, pageSize, filters),
    staleTime: 30 * 1000, // 30 seconds (audit logs change frequently)
  });
};

export const useCreatePatient = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreatePatientRequest) => patientService.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["patients"] });
    },
  });
};

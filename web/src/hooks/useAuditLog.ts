export const useAuditLog = () => {
  const logEvent = async (
    resourceType: string,
    resourceId: string,
    action: string,
    details?: any
  ) => {
    // Audit logging temporarily disabled during migration
    console.log("Audit Log (Simulated):", { resourceType, resourceId, action, details });
  };

  return { logEvent };
};

export const logAuditEvent = async (
  resourceType: string,
  resourceId: string,
  action: string,
  details?: any
) => {
  // Audit logging temporarily disabled during migration
  console.log("Audit Log (Simulated):", { resourceType, resourceId, action, details });
};

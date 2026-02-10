// Payment Status Enum
export const PaymentStatus = {
  PENDING: 'pending',
  APPROVED: 'approved',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled',
  FAILED: 'failed'
};

// PSP Transaction Status
export const TransactionStatus = {
  SUCCESSFUL: 0,
  IN_PROGRESS: 1,
  FAILED: 2,
  ERROR: 3
};

export const getStatusDisplay = (status) => {
  switch(status?.toLowerCase()) {
    case PaymentStatus.PENDING:
      return { text: 'Pending', color: '#FFA500', icon: '⏳' };
    case PaymentStatus.APPROVED:
      return { text: 'Approved', color: '#4CAF50', icon: '✓' };
    case PaymentStatus.COMPLETED:
      return { text: 'Completed', color: '#4CAF50', icon: '✓' };
    case PaymentStatus.CANCELLED:
      return { text: 'Cancelled', color: '#f44336', icon: '✗' };
    case PaymentStatus.FAILED:
      return { text: 'Failed', color: '#f44336', icon: '✗' };
    default:
      return { text: 'Unknown', color: '#757575', icon: '?' };
  }
};

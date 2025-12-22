package com.bankflow.customer.dto;

import com.bankflow.customer.model.KycDocument;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class KycDocumentResponse {
    private UUID id;
    private UUID customerId;
    private KycDocument.DocumentType documentType;
    private String documentNumber;
    private String documentUrl;
    private KycDocument.DocumentStatus status;
    private String rejectionReason;
    private LocalDateTime uploadedAt;
    private LocalDateTime verifiedAt;
}

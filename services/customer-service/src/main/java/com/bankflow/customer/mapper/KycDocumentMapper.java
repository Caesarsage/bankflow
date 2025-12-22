package com.bankflow.customer.mapper;

import com.bankflow.customer.dto.KycDocumentResponse;
import com.bankflow.customer.model.KycDocument;
import org.springframework.stereotype.Component;

@Component
public class KycDocumentMapper {

    public KycDocumentResponse toResponse(KycDocument document) {
        return KycDocumentResponse.builder()
                .id(document.getId())
                .customerId(document.getCustomer().getId())
                .documentType(document.getDocumentType())
                .documentNumber(document.getDocumentNumber())
                .documentUrl(document.getDocumentUrl())
                .status(document.getStatus())
                .rejectionReason(document.getRejectionReason())
                .uploadedAt(document.getUploadedAt())
                .verifiedAt(document.getVerifiedAt())
                .build();
    }
}
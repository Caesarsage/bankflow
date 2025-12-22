package com.bankflow.customer.service;

import com.bankflow.customer.dto.KycDocumentRequest;
import com.bankflow.customer.dto.KycDocumentResponse;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;
import java.util.UUID;

public interface KycDocumentService {
    public KycDocumentResponse uploadDocument(
            UUID customerId,
            KycDocumentRequest request,
            MultipartFile file);
    public List<KycDocumentResponse> getCustomerDocuments(UUID customerId);
    public KycDocumentResponse getDocumentById(UUID documentId);
    public void verifyDocument(UUID documentId, UUID verifiedBy);
    public void rejectDocument(UUID documentId, String reason, UUID rejectedBy);
    public void deleteDocument(UUID documentId);
}

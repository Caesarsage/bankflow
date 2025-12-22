package com.bankflow.customer.controller;

import com.bankflow.customer.dto.KycDocumentRequest;
import com.bankflow.customer.dto.KycDocumentResponse;
import com.bankflow.customer.service.KycDocumentService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/v1/customers/{customerId}/documents")
@RequiredArgsConstructor
@Slf4j
public class KycDocumentController {

    private final KycDocumentService documentService;

    @PostMapping(consumes = MediaType.MULTIPART_FORM_DATA_VALUE)
    public ResponseEntity<KycDocumentResponse> uploadDocument(
            @PathVariable UUID customerId,
            @RequestParam("file") MultipartFile file,
            @RequestParam("documentType") String documentType,
            @RequestParam(value = "documentNumber", required = false) String documentNumber) {

        log.info("Uploading document for customer: {}", customerId);

        KycDocumentRequest request = KycDocumentRequest.builder()
                .documentType(com.bankflow.customer.model.KycDocument.DocumentType.valueOf(documentType))
                .documentNumber(documentNumber)
                .build();

        KycDocumentResponse response = documentService.uploadDocument(customerId, request, file);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping
    public ResponseEntity<List<KycDocumentResponse>> getCustomerDocuments(@PathVariable UUID customerId) {
        log.info("Getting documents for customer: {}", customerId);
        List<KycDocumentResponse> responses = documentService.getCustomerDocuments(customerId);
        return ResponseEntity.ok(responses);
    }

    @GetMapping("/{documentId}")
    public ResponseEntity<KycDocumentResponse> getDocumentById(
            @PathVariable UUID customerId,
            @PathVariable UUID documentId) {
        log.info("Getting document: {}", documentId);
        KycDocumentResponse response = documentService.getDocumentById(documentId);
        return ResponseEntity.ok(response);
    }

    @PatchMapping("/{documentId}/verify")
    public ResponseEntity<Void> verifyDocument(
            @PathVariable UUID customerId,
            @PathVariable UUID documentId,
            @RequestParam UUID verifiedBy) {
        log.info("Verifying document: {}", documentId);
        documentService.verifyDocument(documentId, verifiedBy);
        return ResponseEntity.noContent().build();
    }

    @PatchMapping("/{documentId}/reject")
    public ResponseEntity<Void> rejectDocument(
            @PathVariable UUID customerId,
            @PathVariable UUID documentId,
            @RequestParam String reason,
            @RequestParam UUID rejectedBy) {
        log.info("Rejecting document: {}", documentId);
        documentService.rejectDocument(documentId, reason, rejectedBy);
        return ResponseEntity.noContent().build();
    }

    @DeleteMapping("/{documentId}")
    public ResponseEntity<Void> deleteDocument(
            @PathVariable UUID customerId,
            @PathVariable UUID documentId) {
        log.info("Deleting document: {}", documentId);
        documentService.deleteDocument(documentId);
        return ResponseEntity.noContent().build();
    }
}
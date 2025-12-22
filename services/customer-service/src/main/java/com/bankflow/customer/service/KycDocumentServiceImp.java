package com.bankflow.customer.service;

import com.bankflow.customer.dto.KycDocumentRequest;
import com.bankflow.customer.dto.KycDocumentResponse;
import com.bankflow.customer.exception.CustomerNotFoundException;
import com.bankflow.customer.exception.DocumentNotFoundException;
import com.bankflow.customer.mapper.KycDocumentMapper;
import com.bankflow.customer.model.Customer;
import com.bankflow.customer.model.KycDocument;
import com.bankflow.customer.repository.CustomerRepository;
import com.bankflow.customer.repository.KycDocumentRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class KycDocumentServiceImp implements KycDocumentService {

    private final KycDocumentRepository documentRepository;
    private final CustomerRepository customerRepository;
    private final KycDocumentMapper documentMapper;
    private final StorageService storageService;
    private final CustomerService customerService;

    public KycDocumentResponse uploadDocument(
            UUID customerId,
            KycDocumentRequest request,
            MultipartFile file) {

        log.info("Uploading document for customer: {}", customerId);

        Customer customer = customerRepository.findById(customerId)
                .orElseThrow(() -> new CustomerNotFoundException("Customer not found: " + customerId));

        // Upload file to storage
        String documentUrl = storageService.uploadFile(file, customerId, request.getDocumentType());

        // Create document entity
        KycDocument document = KycDocument.builder()
                .customer(customer)
                .documentType(request.getDocumentType())
                .documentNumber(request.getDocumentNumber())
                .documentUrl(documentUrl)
                .status(KycDocument.DocumentStatus.PENDING)
                .build();

        KycDocument savedDocument = documentRepository.save(document);
        log.info("Document uploaded with ID: {}", savedDocument.getId());

        // Update customer KYC status to IN_REVIEW if still PENDING
        if (customer.getKycStatus() == Customer.KycStatus.PENDING) {
            customerService.updateKycStatus(customerId, Customer.KycStatus.IN_REVIEW);
        }

        return documentMapper.toResponse(savedDocument);
    }

    @Transactional(readOnly = true)
    public List<KycDocumentResponse> getCustomerDocuments(UUID customerId) {
        List<KycDocument> documents = documentRepository.findByCustomerId(customerId);
        return documents.stream()
                .map(documentMapper::toResponse)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public KycDocumentResponse getDocumentById(UUID documentId) {
        KycDocument document = documentRepository.findById(documentId)
                .orElseThrow(() -> new DocumentNotFoundException("Document not found: " + documentId));

        return documentMapper.toResponse(document);
    }

    public void verifyDocument(UUID documentId, UUID verifiedBy) {
        log.info("Verifying document: {}", documentId);

        KycDocument document = documentRepository.findById(documentId)
                .orElseThrow(() -> new DocumentNotFoundException("Document not found: " + documentId));

        document.setStatus(KycDocument.DocumentStatus.VERIFIED);
        document.setVerifiedBy(verifiedBy);
        document.setVerifiedAt(java.time.LocalDateTime.now());

        documentRepository.save(document);
        log.info("Document verified: {}", documentId);

        // Check if all documents are verified
        checkAndUpdateKycStatus(document.getCustomer().getId());
    }

    public void rejectDocument(UUID documentId, String reason, UUID rejectedBy) {
        log.info("Rejecting document: {}", documentId);

        KycDocument document = documentRepository.findById(documentId)
                .orElseThrow(() -> new DocumentNotFoundException("Document not found: " + documentId));

        document.setStatus(KycDocument.DocumentStatus.REJECTED);
        document.setRejectionReason(reason);
        document.setVerifiedBy(rejectedBy);
        document.setVerifiedAt(java.time.LocalDateTime.now());

        documentRepository.save(document);
        log.info("Document rejected: {}", documentId);

        // Update customer KYC status to REJECTED
        customerService.updateKycStatus(document.getCustomer().getId(), Customer.KycStatus.REJECTED);
    }

    public void deleteDocument(UUID documentId) {
        log.info("Deleting document: {}", documentId);

        KycDocument document = documentRepository.findById(documentId)
                .orElseThrow(() -> new DocumentNotFoundException("Document not found: " + documentId));

        // Delete from storage
        storageService.deleteFile(document.getDocumentUrl());

        // Delete from database
        documentRepository.delete(document);
        log.info("Document deleted: {}", documentId);
    }

    private void checkAndUpdateKycStatus(UUID customerId) {
        long totalDocuments = documentRepository.countByCustomerId(customerId);
        long verifiedDocuments = documentRepository.countByCustomerIdAndStatus(customerId, KycDocument.DocumentStatus.VERIFIED);
        // Require at least 2 documents (ID + Proof of Address)
        if (totalDocuments >= 2 && verifiedDocuments == totalDocuments) {
            customerService.updateKycStatus(customerId, Customer.KycStatus.APPROVED);
        }
    }



}
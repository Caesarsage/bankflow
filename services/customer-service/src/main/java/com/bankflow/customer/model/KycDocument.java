package com.bankflow.customer.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;

import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "kyc_documents")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class KycDocument {

    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private UUID id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "customer_id", nullable = false)
    private Customer customer;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 50)
    private DocumentType documentType;

    @Column(length = 100)
    private String documentNumber;

    @Column(nullable = false, length = 500)
    private String documentUrl;

    @Enumerated(EnumType.STRING)
    @Column(length = 50)
    @Builder.Default
    private DocumentStatus status = DocumentStatus.PENDING;

    private UUID verifiedBy;

    @Column(columnDefinition = "TEXT")
    private String rejectionReason;

    @CreationTimestamp
    @Column(nullable = false, updatable = false)
    private LocalDateTime uploadedAt;

    private LocalDateTime verifiedAt;

    public enum DocumentType {
        DRIVERS_LICENSE,
        PASSPORT,
        NATIONAL_ID,
        PROOF_OF_ADDRESS,
        SSN_CARD,
        BIRTH_CERTIFICATE,
        OTHER
    }

    public enum DocumentStatus {
        PENDING,
        VERIFIED,
        REJECTED,
        EXPIRED
    }
}
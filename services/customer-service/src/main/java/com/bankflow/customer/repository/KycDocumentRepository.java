package com.bankflow.customer.repository;

import com.bankflow.customer.model.KycDocument;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface KycDocumentRepository extends JpaRepository<KycDocument, UUID> {

    List<KycDocument> findByCustomerId(UUID customerId);

    List<KycDocument> findByCustomerIdAndStatus(UUID customerId, KycDocument.DocumentStatus status);

    long countByCustomerIdAndStatus(UUID customerId, KycDocument.DocumentStatus status);

    long countByCustomerId(UUID customerId);
}
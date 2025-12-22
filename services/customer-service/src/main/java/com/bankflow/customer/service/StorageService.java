package com.bankflow.customer.service;

import com.bankflow.customer.model.KycDocument;
import org.springframework.web.multipart.MultipartFile;

import java.util.UUID;

public interface StorageService {
    public String uploadFile(MultipartFile file, UUID customerId, KycDocument.DocumentType documentType);
    public void deleteFile(String fileUrl);
}

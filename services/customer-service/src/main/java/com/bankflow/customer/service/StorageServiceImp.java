package com.bankflow.customer.service;

import com.bankflow.customer.exception.StorageException;
import com.bankflow.customer.model.KycDocument;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.io.FilenameUtils;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.util.UUID;

@Service
@Slf4j
public class StorageServiceImp implements StorageService {

    @Value("${app.storage.type:local}")
    private String storageType;

    @Value("${app.storage.local.base-path:/tmp/bankflow/documents}")
    private String basePath;

    public String uploadFile(MultipartFile file, UUID customerId, KycDocument.DocumentType documentType) {
        try {
            // Validate file
            if (file.isEmpty()) {
                throw new StorageException("Cannot upload empty file");
            }

            // Generate unique filename
            String originalFilename = file.getOriginalFilename();
            String extension = FilenameUtils.getExtension(originalFilename);
            String filename = String.format("%s_%s_%s.%s",
                    customerId,
                    documentType.name(),
                    UUID.randomUUID(),
                    extension);

            // Create directory if not exists
            Path uploadPath = Paths.get(basePath, customerId.toString());
            Files.createDirectories(uploadPath);

            // Save file
            Path filePath = uploadPath.resolve(filename);
            Files.copy(file.getInputStream(), filePath, StandardCopyOption.REPLACE_EXISTING);

            log.info("File uploaded successfully: {}", filename);
            return filePath.toString();

        } catch (IOException e) {
            log.error("Failed to upload file", e);
            throw new StorageException("Failed to upload file: " + e.getMessage());
        }
    }

    public void deleteFile(String fileUrl) {
        try {
            Path filePath = Paths.get(fileUrl);
            Files.deleteIfExists(filePath);
            log.info("File deleted: {}", fileUrl);
        } catch (IOException e) {
            log.error("Failed to delete file: {}", fileUrl, e);
            throw new StorageException("Failed to delete file: " + e.getMessage());
        }
    }
}
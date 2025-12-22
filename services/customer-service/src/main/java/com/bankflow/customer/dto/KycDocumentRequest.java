package com.bankflow.customer.dto;

import com.bankflow.customer.model.KycDocument;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class KycDocumentRequest {

    @NotNull(message = "Document type is required")
    private KycDocument.DocumentType documentType;

    private String documentNumber;
}

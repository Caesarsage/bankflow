package com.bankflow.customer.dto;

import com.bankflow.customer.model.Customer;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CustomerResponse {

    private UUID id;
    private UUID userId;
    private String firstName;
    private String lastName;
    private LocalDate dateOfBirth;
    private String addressLine1;
    private String addressLine2;
    private String city;
    private String state;
    private String zipCode;
    private String country;
    private Customer.KycStatus kycStatus;
    private LocalDateTime kycVerifiedAt;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    // Note: SSN is not included in response for security
}
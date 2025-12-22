package com.bankflow.customer.event;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CustomerEvent {
    private String eventId;
    private String eventType;
    private UUID customerId;
    private UUID userId;
    private String kycStatus;
    private LocalDateTime timestamp;
}
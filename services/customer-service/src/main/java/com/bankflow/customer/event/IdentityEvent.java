package com.bankflow.customer.event;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.JsonNode;
import lombok.Data;

import java.time.Instant;

@Data
public class IdentityEvent {

    @JsonProperty("event_id") // JSON event comes as snake_case
    private String eventId;

    @JsonProperty("event_type") // JSON event comes as snake_case
    private String eventType;

    @JsonProperty("user_id") // JSON event comes as snake_case
    private String userId;

    private String email;
    private JsonNode data;
    private Instant timestamp;
}

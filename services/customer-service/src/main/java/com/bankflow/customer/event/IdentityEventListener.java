package com.bankflow.customer.event;

import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.support.KafkaHeaders;
import org.springframework.messaging.handler.annotation.Header;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.stereotype.Component;

@Component
@RequiredArgsConstructor
@Slf4j
public class IdentityEventListener {

    private final ObjectMapper objectMapper;

    @KafkaListener(topics = "identity-events", groupId = "customer-service")
    public void handleIdentityEvent(@Payload String message) {
        try {
            IdentityEvent event =
                    objectMapper.readValue(message, IdentityEvent.class);

            log.info("Received event type: {}", event.getEventType());

            switch (event.getEventType()) {

                case "user.registered" -> {
                    UserRegisteredEvent registeredEvent =
                            objectMapper.convertValue(event, UserRegisteredEvent.class);

                    log.info("User registered: {}", registeredEvent.getUserId());
                    // create customer profile
                }

                case "user.logged_in" -> {
                    log.info("User logged in: {}", event.getUserId());
                }

                default -> log.warn("Unhandled event type: {}", event.getEventType());
            }

        } catch (Exception e) {
            log.error("Failed to process identity event", e);
        }
    }
}

package com.bankflow.customer.service;

import com.bankflow.customer.dto.CustomerRequest;
import com.bankflow.customer.dto.CustomerResponse;
import com.bankflow.customer.event.CustomerEvent;
import com.bankflow.customer.exception.CustomerAlreadyExistsException;
import com.bankflow.customer.exception.CustomerNotFoundException;
import com.bankflow.customer.mapper.CustomerMapper;
import com.bankflow.customer.model.Customer;
import com.bankflow.customer.repository.CustomerRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class CustomerServiceImp implements CustomerService {

    private final CustomerRepository customerRepository;
    private final CustomerMapper customerMapper;
    private final KafkaTemplate<String, Object> kafkaTemplate;
    private final EncryptionService encryptionService;

    private static final String CUSTOMER_EVENTS_TOPIC = "customer-events";

    public CustomerResponse createCustomer(CustomerRequest request) {
        log.info("Creating customer for user: {}", request.getUserId());

        // Check if customer already exists
        if (customerRepository.existsByUserId(request.getUserId())) {
            throw new CustomerAlreadyExistsException("Customer already exists for user: " + request.getUserId());
        }

        // Encrypt SSN if provided
        String encryptedSsn = null;
        if (request.getSsn() != null && !request.getSsn().isEmpty()) {
            encryptedSsn = encryptionService.encrypt(request.getSsn());
        }

        // Create customer entity
        Customer customer = customerMapper.toEntity(request);
        customer.setSsnEncrypted(encryptedSsn);
        customer.setKycStatus(Customer.KycStatus.PENDING);

        // Save customer
        Customer savedCustomer = customerRepository.save(customer);
        log.info("Customer created with ID: {}", savedCustomer.getId());

        // Publish event
        publishCustomerEvent("customer.created", savedCustomer);

        return customerMapper.toResponse(savedCustomer);
    }

    @Transactional(readOnly = true)
    public CustomerResponse getCustomerById(UUID customerId) {
        Customer customer = customerRepository.findById(customerId)
                .orElseThrow(() -> new CustomerNotFoundException("Customer not found: " + customerId));

        return customerMapper.toResponse(customer);
    }

    @Transactional(readOnly = true)
    public CustomerResponse getCustomerByUserId(UUID userId) {
        Customer customer = customerRepository.findByUserId(userId)
                .orElseThrow(() -> new CustomerNotFoundException("Customer not found for user: " + userId));

        return customerMapper.toResponse(customer);
    }

    public CustomerResponse updateCustomer(UUID customerId, CustomerRequest request) {
        log.info("Updating customer: {}", customerId);

        Customer customer = customerRepository.findById(customerId)
                .orElseThrow(() -> new CustomerNotFoundException("Customer not found: " + customerId));

        // Update fields
        customer.setFirstName(request.getFirstName());
        customer.setLastName(request.getLastName());
        customer.setDateOfBirth(request.getDateOfBirth());
        customer.setAddressLine1(request.getAddressLine1());
        customer.setAddressLine2(request.getAddressLine2());
        customer.setCity(request.getCity());
        customer.setState(request.getState());
        customer.setZipCode(request.getZipCode());
        customer.setCountry(request.getCountry());

        // Update SSN if provided
        if (request.getSsn() != null && !request.getSsn().isEmpty()) {
            customer.setSsnEncrypted(encryptionService.encrypt(request.getSsn()));
        }

        Customer updatedCustomer = customerRepository.save(customer);
        log.info("Customer updated: {}", customerId);

        // Publish event
        publishCustomerEvent("customer.updated", updatedCustomer);

        return customerMapper.toResponse(updatedCustomer);
    }

    @Transactional(readOnly = true)
    public List<CustomerResponse> searchCustomers(String query) {
        List<Customer> customers = customerRepository.searchCustomers(query);
        return customers.stream()
                .map(customerMapper::toResponse)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public List<CustomerResponse> getCustomersByKycStatus(Customer.KycStatus status) {
        List<Customer> customers = customerRepository.findByKycStatus(status);
        return customers.stream()
                .map(customerMapper::toResponse)
                .collect(Collectors.toList());
    }

    public void updateKycStatus(UUID customerId, Customer.KycStatus newStatus) {
        log.info("Updating KYC status for customer {} to {}", customerId, newStatus);

        Customer customer = customerRepository.findById(customerId)
                .orElseThrow(() -> new CustomerNotFoundException("Customer not found: " + customerId));

        Customer.KycStatus oldStatus = customer.getKycStatus();
        customer.setKycStatus(newStatus);

        if (newStatus == Customer.KycStatus.APPROVED) {
            customer.setKycVerifiedAt(LocalDateTime.now());
        }

        customerRepository.save(customer);
        log.info("KYC status updated from {} to {}", oldStatus, newStatus);

        // Publish event
        if (newStatus == Customer.KycStatus.APPROVED) {
            publishCustomerEvent("kyc.approved", customer);
        } else if (newStatus == Customer.KycStatus.REJECTED) {
            publishCustomerEvent("kyc.rejected", customer);
        }
    }

    private void publishCustomerEvent(String eventType, Customer customer) {
        try {
            CustomerEvent event = CustomerEvent.builder()
                    .eventId(UUID.randomUUID().toString())
                    .eventType(eventType)
                    .customerId(customer.getId())
                    .userId(customer.getUserId())
                    .kycStatus(customer.getKycStatus().name())
                    .timestamp(LocalDateTime.now())
                    .build();

            kafkaTemplate.send(CUSTOMER_EVENTS_TOPIC, customer.getId().toString(), event);
            log.info("Published event: {} for customer: {}", eventType, customer.getId());
        } catch (Exception e) {
            log.error("Failed to publish event: {}", eventType, e);
        }
    }
}

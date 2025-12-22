package com.bankflow.customer.controller;

import com.bankflow.customer.dto.CustomerRequest;
import com.bankflow.customer.dto.CustomerResponse;
import com.bankflow.customer.model.Customer;
import com.bankflow.customer.service.CustomerService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/v1/customers")
@RequiredArgsConstructor
@Slf4j
public class CustomerController {

    private final CustomerService customerService;

    @PostMapping
    public ResponseEntity<CustomerResponse> createCustomer(@Valid @RequestBody CustomerRequest request) {
        log.info("Creating customer for user: {}", request.getUserId());
        CustomerResponse response = customerService.createCustomer(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping("/{id}")
    public ResponseEntity<CustomerResponse> getCustomerById(@PathVariable UUID id) {
        log.info("Getting customer by ID: {}", id);
        CustomerResponse response = customerService.getCustomerById(id);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/user/{userId}")
    public ResponseEntity<CustomerResponse> getCustomerByUserId(@PathVariable UUID userId) {
        log.info("Getting customer by user ID: {}", userId);
        CustomerResponse response = customerService.getCustomerByUserId(userId);
        return ResponseEntity.ok(response);
    }

    @PutMapping("/{id}")
    public ResponseEntity<CustomerResponse> updateCustomer(
            @PathVariable UUID id,
            @Valid @RequestBody CustomerRequest request) {
        log.info("Updating customer: {}", id);
        CustomerResponse response = customerService.updateCustomer(id, request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/search")
    public ResponseEntity<List<CustomerResponse>> searchCustomers(@RequestParam String query) {
        log.info("Searching customers with query: {}", query);
        List<CustomerResponse> responses = customerService.searchCustomers(query);
        return ResponseEntity.ok(responses);
    }

    @GetMapping("/kyc-status/{status}")
    public ResponseEntity<List<CustomerResponse>> getCustomersByKycStatus(
            @PathVariable Customer.KycStatus status) {
        log.info("Getting customers by KYC status: {}", status);
        List<CustomerResponse> responses = customerService.getCustomersByKycStatus(status);
        return ResponseEntity.ok(responses);
    }

    @PatchMapping("/{id}/kyc-status")
    public ResponseEntity<Void> updateKycStatus(
            @PathVariable UUID id,
            @RequestParam Customer.KycStatus status) {
        log.info("Updating KYC status for customer {} to {}", id, status);
        customerService.updateKycStatus(id, status);
        return ResponseEntity.noContent().build();
    }
}
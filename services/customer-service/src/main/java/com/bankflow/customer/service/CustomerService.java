package com.bankflow.customer.service;

import com.bankflow.customer.dto.CustomerRequest;
import com.bankflow.customer.dto.CustomerResponse;
import com.bankflow.customer.model.Customer;

import java.util.List;
import java.util.UUID;

public interface CustomerService {
    public CustomerResponse createCustomer(CustomerRequest request);
    public CustomerResponse getCustomerById(UUID customerId);
    public CustomerResponse getCustomerByUserId(UUID userId);
    public CustomerResponse updateCustomer(UUID customerId, CustomerRequest request);
    public List<CustomerResponse> searchCustomers(String query);
    public List<CustomerResponse> getCustomersByKycStatus(Customer.KycStatus status);
    public void updateKycStatus(UUID customerId, Customer.KycStatus newStatus);
}

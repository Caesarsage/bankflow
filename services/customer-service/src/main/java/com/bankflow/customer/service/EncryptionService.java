package com.bankflow.customer.service;

import org.springframework.beans.factory.annotation.Value;

public interface EncryptionService {
    public String encrypt(String value);
    public String decrypt(String encryptedValue);
}

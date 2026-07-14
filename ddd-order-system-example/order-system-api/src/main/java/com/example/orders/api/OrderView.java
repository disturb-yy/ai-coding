package com.example.orders.api;

import java.math.BigDecimal;

public record OrderView(String id, String customerId, BigDecimal totalAmount, String status) {
}

package com.example.orders.api;

import java.math.BigDecimal;

/** 对外稳定写入契约，不承载领域行为。 */
public record CreateOrderCommand(String customerId, BigDecimal totalAmount) {
}

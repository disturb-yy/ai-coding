package com.example.orders.domain;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import java.math.BigDecimal;
import org.junit.jupiter.api.Test;

class OrderTest {

    @Test
    void pendingOrderCanBeConfirmed() {
        var order = Order.create(OrderId.newId(), "customer-1", new Money(new BigDecimal("12.50")));

        order.confirm();

        assertThat(order.status()).isEqualTo(OrderStatus.CONFIRMED);
    }

    @Test
    void cancelledOrderCannotBeConfirmed() {
        var order = Order.create(OrderId.newId(), "customer-1", new Money(new BigDecimal("12.50")));
        order.cancel("客户主动取消");

        assertThatThrownBy(order::confirm)
                .isInstanceOf(IllegalStateException.class)
                .hasMessage("只有待确认订单可以确认");
    }
}

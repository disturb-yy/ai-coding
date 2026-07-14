package com.example.orders.app;

import static org.assertj.core.api.Assertions.assertThat;

import com.example.orders.api.CreateOrderCommand;
import com.example.orders.domain.Order;
import com.example.orders.domain.OrderId;
import com.example.orders.domain.OrderRepository;
import java.math.BigDecimal;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import org.junit.jupiter.api.Test;

class OrderCommandServiceImplTest {

    @Test
    void createThenConfirmDelegatesStateChangeToAggregate() {
        var service = new OrderCommandServiceImpl(new InMemoryTestRepository());
        var created = service.create(new CreateOrderCommand("customer-1", new BigDecimal("20.00")));

        var confirmed = service.confirm(created.id());

        assertThat(confirmed.status()).isEqualTo("CONFIRMED");
    }

    private static final class InMemoryTestRepository implements OrderRepository {
        private final Map<OrderId, Order> orders = new HashMap<>();

        @Override
        public Optional<Order> findById(OrderId id) {
            return Optional.ofNullable(orders.get(id));
        }

        @Override
        public void save(Order order) {
            orders.put(order.id(), order);
        }
    }
}

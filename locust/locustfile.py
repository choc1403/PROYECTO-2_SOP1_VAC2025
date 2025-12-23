from locust import HttpUser, TaskSet, task, between
import random

class MyTasks(TaskSet):

    @task(1)
    def enviar_venta(self):
        categorias = ["Electronica", "Ropa", "Hogar", "Belleza"]

        venta_data = {
            "categoria": random.choice(categorias),        # Seleccion random de categorias
            "producto_id": f"P{random.randint(1,10)}",     # Random de productos entre 1 y 10
            "precio": random.uniform(40, 1000),            # Random de precio permitiendo decimales entre 40 y 1000
            "cantidad_vendida": random.randint(1, 20)      # Random de cantidades entre 1 y 20
        }

        headers = {"Content-Type": "application/json"}
        self.client.post("/venta", json=venta_data, headers=headers)

class WebsiteUser(HttpUser):
    tasks = [MyTasks]
    wait_time = between(1, 5)

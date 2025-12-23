# Locustfile.py
from locust import HttpUser, TaskSet, task, between
import random
import json

class MyTasks(TaskSet):
    
    @task(1)
    def engineering(self):
        categorias = ["Electronica", "Ropa", "Hogar", "Belleza"]
    
        # Datos de venta
        venta_data = {
            "categoria": random.choices(categorias),  # Random Categoria
            "producto_id": random.randint(1,10),  # Random de producto_id de 1 y 10
            "precio": random.randint(40, 1000),  # Precio random entre 40 y 80
            "cantidad_vendida": random.randint(1,20)  # Cantidad random entre 1 y 20
        }
        
        # Envio de JSON hacia route como POST
        headers = {'Content-Type': 'application/json'}
        self.client.post("/venta", json=venta_data, headers=headers)

class WebsiteUser(HttpUser):
    tasks = [MyTasks]
    wait_time = between(1, 5)  # Tiempo de espera entre tareas entre 1 y 5 segundos
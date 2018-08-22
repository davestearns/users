provider "aws" {
}

data "aws_vpc" "default-vpc" {
    default = true
}

data "aws_subnet_ids" "default-vpc-subnets" {
    vpc_id = "${data.aws_vpc.default-vpc.id}"
}

# users table
resource "aws_dynamodb_table" "users-table" {
    name = "users"
    read_capacity = 5
    write_capacity = 5
    hash_key = "userName"

    attribute {
        name = "userName"
        type = "S"
    }
}

# session cache
resource "aws_security_group" "session-cache-sg" {
    name = "session-cache-sg"
    
    ingress {
        protocol = "tcp"
        from_port = 6379
        to_port = 6379
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_cluster" "session-cache" {
    cluster_id = "session-cache"
    engine = "redis"
    node_type = "cache.m3.medium"
    num_cache_nodes = 1
    security_group_ids = ["${aws_security_group.session-cache-sg.id}"]
}

# IAM role for ECS tasks
data "aws_iam_policy_document" "user-service-role-policy" {
    statement {
        actions = ["sts:AssumeRole"]
        principals {
            type = "Service"
            identifiers = ["ecs-tasks.amazonaws.com"]
        }
    }
}

resource "aws_iam_role" "user-service-role" {
    name = "user-service-role"
    assume_role_policy = "${data.aws_iam_policy_document.user-service-role-policy.json}"
}

resource "aws_iam_role_policy_attachment" "user-service-role-attach-dynamodb" {
    role = "${aws_iam_role.user-service-role.name}"
    policy_arn = "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess"
}

resource "aws_iam_role_policy_attachment" "user-service-role-attach-secrets" {
    role = "${aws_iam_role.user-service-role.name}"
    policy_arn = "arn:aws:iam::aws:policy/SecretsManagerReadWrite"
}

# ECS Task Definition
resource "aws_ecs_task_definition" "users-taskdef" {
    family = "users"
    # NOTE: you must include hostPort and protocol in the port mappings
    # even though they aren't required, in order to avoid re-creating the
    # task definition every time you run terraform apply
    container_definitions = <<EOF
[{
    "name": "users",
    "image": "davestearns/userservice",
    "portMappings": [{"containerPort": 80, "hostPort": 80, "protocol": "tcp"}],
    "environment": [
        {
            "name": "REDIS_ADDR", 
            "value": "${aws_elasticache_cluster.session-cache.cache_nodes.0.address}:${aws_elasticache_cluster.session-cache.cache_nodes.0.port}"
        }
    ]
}]
EOF
    task_role_arn = "${aws_iam_role.user-service-role.arn}"
    execution_role_arn = "arn:aws:iam::008944543045:role/ecsTaskExecutionRole"
    network_mode = "awsvpc"
    requires_compatibilities = ["FARGATE"]
    cpu = 512
    memory = 1024
}

# ECS Cluster
resource "aws_ecs_cluster" "user-service-cluster" {
    name = "user-service-cluster"
}

# Load Balancer

data "aws_acm_certificate" "api-cert" {
    domain = "api.info441.info"
    most_recent = true 
    statuses = ["ISSUED"]
}

resource "aws_security_group" "user-service-lb-sg" {
    name = "user-service-lb-sg"
    
    ingress {
        protocol = "tcp"
        from_port = 443
        to_port = 443
        cidr_blocks = ["0.0.0.0/0"]
    }
    egress {
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_lb" "user-service-lb" {
    name = "user-service-lb"
    internal = false
    load_balancer_type = "application"
    security_groups = ["${aws_security_group.user-service-lb-sg.id}"]
    subnets = ["${data.aws_subnet_ids.default-vpc-subnets.ids}"]   
}

resource "aws_lb_target_group" "user-service-target-group" {
    name = "user-service-target-group"
    target_type = "ip"
    port = 80
    protocol = "HTTP"
    vpc_id = "${data.aws_vpc.default-vpc.id}"
}

resource "aws_lb_listener" "user-service-listener" {
    load_balancer_arn = "${aws_lb.user-service-lb.arn}"
    port = 443
    protocol = "HTTPS"
    certificate_arn = "${data.aws_acm_certificate.api-cert.arn}"
    default_action {
        type = "forward"
        target_group_arn = "${aws_lb_target_group.user-service-target-group.arn}"
    }
}

# ECS Service
resource "aws_security_group" "user-service-sg" {
    name = "user-service-sg"
    
    ingress {
        protocol = "tcp"
        from_port = 80
        to_port = 80
        cidr_blocks = ["0.0.0.0/0"]
    }
    egress {
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_ecs_service" "user-serivce" {
    name = "user-service"
    task_definition = "${aws_ecs_task_definition.users-taskdef.arn}"
    cluster = "${aws_ecs_cluster.user-service-cluster.arn}"
    desired_count = 2
    launch_type = "FARGATE"
    network_configuration = {
        subnets = ["${data.aws_subnet_ids.default-vpc-subnets.ids}"]
        security_groups = ["${aws_security_group.user-service-sg.id}"]
        assign_public_ip = true
    }
    load_balancer {
        target_group_arn = "${aws_lb_target_group.user-service-target-group.arn}"
        container_name = "users"
        container_port = 80
    }
    depends_on = ["aws_lb.user-service-lb"]
}

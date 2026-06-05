
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
DROP TABLE IF EXISTS `addresses`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `addresses` (
  `id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `province` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `city` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `district` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `detail` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `addresses` WRITE;
/*!40000 ALTER TABLE `addresses` DISABLE KEYS */;
INSERT INTO `addresses` VALUES ('a001','u001','Alice','13800000001','广东省','深圳市','南山区','科技园南路18号',1,'2026-06-02 15:20:56.541'),('a002','u001','Alice公司','13800000001','广东省','深圳市','福田区','华强北路188号',0,'2026-06-02 15:20:56.541'),('a003','u002','Bob','13800000002','北京市','北京市','朝阳区','建国路88号',1,'2026-06-02 15:20:56.541'),('a004','u003','Carol','13800000003','上海市','上海市','浦东新区','陆家嘴环路1000号',1,'2026-06-02 15:20:56.541'),('a005','u004','David','13800000004','浙江省','杭州市','西湖区','文三路477号',1,'2026-06-02 15:20:56.541'),('a006','u005','Eve','13800000005','四川省','成都市','高新区','天府大道699号',1,'2026-06-02 15:20:56.541'),('a007','u006','Frank','13800000006','湖北省','武汉市','洪山区','光谷大道55号',1,'2026-06-02 15:20:56.541'),('a008','u007','Grace','13800000007','陕西省','西安市','雁塔区','科技路68号',1,'2026-06-02 15:20:56.541'),('a009','u009','Ivy','13800000009','江苏省','南京市','鼓楼区','中山路100号',1,'2026-06-02 15:20:56.541'),('a010','u010','Jack','13800000010','广东省','广州市','天河区','天河路385号',1,'2026-06-02 15:20:56.541'),('a011','u011','Kate','13800000011','湖南省','长沙市','岳麓区','麓谷大道88号',1,'2026-06-02 15:20:56.541'),('a012','u012','Leo','13800000012','福建省','厦门市','思明区','中山路1号',1,'2026-06-02 15:20:56.541'),('a013','u014','Noah','13800000014','山东省','青岛市','市南区','中山路88号',1,'2026-06-02 15:20:56.541'),('a014','u015','Olivia','13800000015','河南省','郑州市','金水区','金水路180号',1,'2026-06-02 15:20:56.541'),('a015','u002','Bob家','13800000002','北京市','北京市','海淀区','中关村大街1号',0,'2026-06-02 15:20:56.541'),('a016','u003','Carol备用','13800000003','上海市','上海市','徐汇区','漕溪北路168号',0,'2026-06-02 15:20:56.541'),('a017','u004','David公司','13800000004','浙江省','杭州市','余杭区','未来科技城25号',0,'2026-06-02 15:20:56.541'),('a018','u005','Eve公司','13800000005','四川省','成都市','武侯区','武侯大道669号',0,'2026-06-02 15:20:56.541'),('a019','u009','Ivy公司','13800000009','江苏省','苏州市','工业园区','星湖街328号',0,'2026-06-02 15:20:56.541'),('a020','u010','Jack公司','13800000010','广东省','深圳市','龙华区','民治大道1168号',0,'2026-06-02 15:20:56.541');
/*!40000 ALTER TABLE `addresses` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `categories` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `parent_id` int DEFAULT NULL,
  `sort_order` int NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `categories` WRITE;
/*!40000 ALTER TABLE `categories` DISABLE KEYS */;
INSERT INTO `categories` VALUES (1,'电子数码',NULL,1,'2026-06-02 15:20:56.518'),(2,'服装鞋包',NULL,2,'2026-06-02 15:20:56.518'),(3,'食品饮料',NULL,3,'2026-06-02 15:20:56.518'),(4,'家居家电',NULL,4,'2026-06-02 15:20:56.518'),(5,'手机',1,1,'2026-06-02 15:20:56.518'),(6,'笔记本电脑',1,2,'2026-06-02 15:20:56.518'),(7,'耳机音响',1,3,'2026-06-02 15:20:56.518'),(8,'男装',2,1,'2026-06-02 15:20:56.518'),(9,'女装',2,2,'2026-06-02 15:20:56.518'),(10,'零食',3,1,'2026-06-02 15:20:56.518');
/*!40000 ALTER TABLE `categories` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `order_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `order_items` (
  `id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `order_id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `product_id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `product_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `unit_price` decimal(10,2) NOT NULL,
  `quantity` int NOT NULL,
  `subtotal` decimal(10,2) NOT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_order` (`order_id`),
  KEY `idx_product` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `order_items` WRITE;
/*!40000 ALTER TABLE `order_items` DISABLE KEYS */;
INSERT INTO `order_items` VALUES ('oi001','o001','p001','iPhone 15 Pro',8999.00,1,8999.00,'2026-06-02 15:21:34.189'),('oi002','o002','p005','MacBook Pro 14',14999.00,1,14999.00,'2026-06-02 15:21:34.189'),('oi003','o003','p008','Sony WH-1000XM5',2499.00,1,2499.00,'2026-06-02 15:21:34.189'),('oi004','o004','p009','AirPods Pro 2',1899.00,1,1899.00,'2026-06-02 15:21:34.189'),('oi005','o005','p002','华为 Mate 60 Pro',6999.00,1,6999.00,'2026-06-02 15:21:34.189'),('oi006','o006','p012','Levi\'s 501直筒牛仔裤',499.00,1,499.00,'2026-06-02 15:21:34.189'),('oi007','o007','p003','小米14 Ultra',5999.00,1,5999.00,'2026-06-02 15:21:34.189'),('oi008','o008','p013','ZARA印花连衣裙',399.00,1,399.00,'2026-06-02 15:21:34.189'),('oi009','o009','p006','联想ThinkPad X1 Carbon',9999.00,1,9999.00,'2026-06-02 15:21:34.189'),('oi010','o010','p004','OPPO Find X7',4999.00,1,4999.00,'2026-06-02 15:21:34.189'),('oi011','o011','p011','优衣库法兰绒格纹衬衫',199.00,1,199.00,'2026-06-02 15:21:34.189'),('oi012','o012','p010','Bose QuietComfort 45',2199.00,1,2199.00,'2026-06-02 15:21:34.189'),('oi013','o013','p007','华为MateBook X Pro',7999.00,1,7999.00,'2026-06-02 15:21:34.189'),('oi014','o014','p015','三只松鼠零食大礼包',99.00,1,99.00,'2026-06-02 15:21:34.189'),('oi015','o015','p018','Dell XPS 15',12999.00,1,12999.00,'2026-06-02 15:21:34.189'),('oi016','o016','p017','vivo X100 Pro',4499.00,1,4499.00,'2026-06-02 15:21:34.189'),('oi017','o017','p009','AirPods Pro 2',1899.00,1,1899.00,'2026-06-02 15:21:34.189'),('oi018','o018','p016','良品铺子肉干礼盒',129.00,1,129.00,'2026-06-02 15:21:34.189'),('oi019','o019','p019','JBL Charge 5',899.00,1,899.00,'2026-06-02 15:21:34.189'),('oi020','o020','p014','HM针织开衫',149.00,1,149.00,'2026-06-02 15:21:34.189'),('oi021','o021','p002','华为 Mate 60 Pro',6999.00,1,6999.00,'2026-06-02 15:21:34.189'),('oi022','o022','p008','Sony WH-1000XM5',2499.00,1,2499.00,'2026-06-02 15:21:34.189'),('oi023','o023','p001','iPhone 15 Pro',8999.00,1,8999.00,'2026-06-02 15:21:34.189'),('oi024','o024','p020','北面(TNF)抓绒外套',999.00,1,999.00,'2026-06-02 15:21:34.189'),('oi025','o025','p003','小米14 Ultra',5999.00,1,5999.00,'2026-06-02 15:21:34.189'),('oi027','o005','p015','三只松鼠零食大礼包',99.00,2,198.00,'2026-06-02 15:21:34.189'),('oi028','o009','p011','优衣库法兰绒格纹衬衫',199.00,1,199.00,'2026-06-02 15:21:34.189'),('oi029','o010','p016','良品铺子肉干礼盒',129.00,2,258.00,'2026-06-02 15:21:34.189'),('oi031','o002','p009','AirPods Pro 2',1899.00,1,1899.00,'2026-06-02 15:21:34.189'),('oi032','o003','p011','优衣库法兰绒格纹衬衫',199.00,2,398.00,'2026-06-02 15:21:34.189'),('oi033','o007','p015','三只松鼠零食大礼包',99.00,3,297.00,'2026-06-02 15:21:34.189'),('oi034','o012','p014','HM针织开衫',149.00,1,149.00,'2026-06-02 15:21:34.189'),('oi035','o013','p016','良品铺子肉干礼盒',129.00,2,258.00,'2026-06-02 15:21:34.189'),('oi036','o014','p016','良品铺子肉干礼盒',129.00,1,129.00,'2026-06-02 15:21:34.189'),('oi037','o022','p011','优衣库法兰绒格纹衬衫',199.00,1,199.00,'2026-06-02 15:21:34.189'),('oi038','o024','p015','三只松鼠零食大礼包',99.00,2,198.00,'2026-06-02 15:21:34.189'),('oi039','o025','p014','HM针织开衫',149.00,2,298.00,'2026-06-02 15:21:34.189'),('oi040','o023','p019','JBL Charge 5',899.00,1,899.00,'2026-06-02 15:21:34.189');
/*!40000 ALTER TABLE `order_items` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `orders`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `orders` (
  `id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `order_no` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `address_id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `total_amount` decimal(10,2) NOT NULL,
  `paid_amount` decimal(10,2) NOT NULL DEFAULT '0.00',
  `status` enum('pending','paid','shipped','delivered','cancelled','refunded') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending',
  `remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `paid_at` datetime(3) DEFAULT NULL,
  `shipped_at` datetime(3) DEFAULT NULL,
  `delivered_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `order_no` (`order_no`),
  KEY `idx_user` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_order_no` (`order_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `orders` WRITE;
/*!40000 ALTER TABLE `orders` DISABLE KEYS */;
INSERT INTO `orders` VALUES ('o001','ORD20240101001','u001','a001',8999.00,8999.00,'delivered',NULL,'2024-01-01 10:00:00.000','2024-01-02 08:00:00.000','2024-01-04 15:30:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o002','ORD20240102001','u002','a003',14999.00,14999.00,'delivered',NULL,'2024-01-02 11:00:00.000','2024-01-03 09:00:00.000','2024-01-05 14:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o003','ORD20240103001','u003','a004',2499.00,2499.00,'delivered',NULL,'2024-01-03 14:00:00.000','2024-01-04 10:00:00.000','2024-01-06 16:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o004','ORD20240110001','u001','a001',1899.00,1899.00,'delivered',NULL,'2024-01-10 09:30:00.000','2024-01-11 08:00:00.000','2024-01-13 11:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o005','ORD20240115001','u004','a005',6999.00,6999.00,'delivered','请尽快发货','2024-01-15 16:00:00.000','2024-01-16 07:30:00.000','2024-01-18 13:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o006','ORD20240120001','u005','a006',499.00,499.00,'delivered',NULL,'2024-01-20 10:00:00.000','2024-01-21 09:00:00.000','2024-01-23 14:30:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o007','ORD20240201001','u006','a007',5999.00,5999.00,'delivered',NULL,'2024-02-01 11:00:00.000','2024-02-02 08:00:00.000','2024-02-04 15:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o008','ORD20240210001','u007','a008',399.00,399.00,'delivered',NULL,'2024-02-10 14:00:00.000','2024-02-11 09:00:00.000','2024-02-13 16:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o009','ORD20240215001','u002','a003',9999.00,9999.00,'delivered',NULL,'2024-02-15 10:30:00.000','2024-02-16 08:30:00.000','2024-02-18 14:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o010','ORD20240301001','u009','a009',4999.00,4999.00,'delivered',NULL,'2024-03-01 09:00:00.000','2024-03-02 08:00:00.000','2024-03-04 13:30:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o011','ORD20240310001','u010','a010',199.00,199.00,'delivered',NULL,'2024-03-10 11:00:00.000','2024-03-11 09:00:00.000','2024-03-13 15:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o012','ORD20240320001','u011','a011',2199.00,2199.00,'delivered',NULL,'2024-03-20 14:00:00.000','2024-03-21 09:30:00.000','2024-03-23 16:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o013','ORD20240401001','u012','a012',7999.00,7999.00,'shipped',NULL,'2024-04-01 10:00:00.000','2024-04-02 08:00:00.000',NULL,'2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o014','ORD20240410001','u001','a002',99.00,99.00,'delivered',NULL,'2024-04-10 11:30:00.000','2024-04-11 09:00:00.000','2024-04-13 14:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o015','ORD20240415001','u003','a004',12999.00,12999.00,'delivered',NULL,'2024-04-15 09:00:00.000','2024-04-16 08:00:00.000','2024-04-18 15:30:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o016','ORD20240501001','u014','a013',4499.00,4499.00,'shipped',NULL,'2024-05-01 10:00:00.000','2024-05-02 08:00:00.000',NULL,'2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o017','ORD20240510001','u015','a014',1899.00,1899.00,'paid','工作日送货','2024-05-10 14:00:00.000',NULL,NULL,'2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o018','ORD20240515001','u004','a005',129.00,129.00,'paid',NULL,'2024-05-15 11:00:00.000',NULL,NULL,'2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o019','ORD20240520001','u005','a006',899.00,899.00,'paid',NULL,'2024-05-20 09:30:00.000',NULL,NULL,'2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o020','ORD20240601001','u002','a015',149.00,0.00,'cancelled',NULL,NULL,NULL,NULL,'2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o021','ORD20240605001','u006','a007',6999.00,6999.00,'refunded','质量问题退款','2024-06-05 10:00:00.000',NULL,NULL,'2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o022','ORD20240610001','u009','a019',2499.00,2499.00,'delivered',NULL,'2024-06-10 09:00:00.000','2024-06-11 08:00:00.000','2024-06-13 15:00:00.000','2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o023','ORD20240615001','u010','a020',8999.00,0.00,'pending',NULL,NULL,NULL,NULL,'2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o024','ORD20240620001','u011','a011',999.00,999.00,'paid',NULL,'2024-06-20 14:00:00.000',NULL,NULL,'2026-06-02 15:21:34.173','2026-06-02 15:21:34.173'),('o025','ORD20240625001','u012','a012',5999.00,5999.00,'shipped',NULL,'2024-06-25 10:00:00.000','2024-06-26 08:00:00.000',NULL,'2026-06-02 15:21:34.173','2026-06-02 15:21:34.173');
/*!40000 ALTER TABLE `orders` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `products`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `products` (
  `id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `category_id` int NOT NULL,
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `price` decimal(10,2) NOT NULL,
  `stock` int NOT NULL DEFAULT '0',
  `status` enum('on_sale','off_sale','out_of_stock') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'on_sale',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_category` (`category_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `products` WRITE;
/*!40000 ALTER TABLE `products` DISABLE KEYS */;
INSERT INTO `products` VALUES ('p001',5,'iPhone 15 Pro','苹果旗舰手机，A17芯片，钛金属边框',8999.00,120,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p002',5,'华为 Mate 60 Pro','麒麟9000S芯片，卫星通话',6999.00,80,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p003',5,'小米14 Ultra','徕卡摄影系统，骁龙8Gen3',5999.00,200,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p004',5,'OPPO Find X7','哈苏影像系统，天玑9300',4999.00,150,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p005',6,'MacBook Pro 14','M3 Pro芯片，14英寸Liquid视网膜屏',14999.00,50,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p006',6,'联想ThinkPad X1 Carbon','Intel i7，16GB内存，轻薄商务本',9999.00,60,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p007',6,'华为MateBook X Pro','13英寸，3K触控屏',7999.00,70,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p008',7,'Sony WH-1000XM5','业界顶级主动降噪耳机',2499.00,300,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p009',7,'AirPods Pro 2','苹果无线降噪耳机，H2芯片',1899.00,500,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p010',7,'Bose QuietComfort 45','舒适佩戴，优质降噪',2199.00,100,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p011',8,'优衣库法兰绒格纹衬衫','男款，多色可选',199.00,1000,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p012',8,'Levi\'s 501直筒牛仔裤','经典款，100%棉',499.00,600,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p013',9,'ZARA印花连衣裙','女款，夏季新款',399.00,400,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p014',9,'HM针织开衫','女款，多色',149.00,800,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p015',10,'三只松鼠零食大礼包','含坚果/饼干/糖果等12种零食',99.00,2000,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p016',10,'良品铺子肉干礼盒','猪肉脯/牛肉干/鸡胸肉组合',129.00,1500,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p017',5,'vivo X100 Pro','蔡司影像，天玑9300',4499.00,180,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p018',6,'Dell XPS 15','OLED屏，RTX 4060',12999.00,40,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p019',7,'JBL Charge 5','防水便携蓝牙音箱',899.00,400,'on_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531'),('p020',8,'北面(TNF)抓绒外套','男款，保暖防风',999.00,300,'off_sale','2026-06-02 15:20:56.531','2026-06-02 15:20:56.531');
/*!40000 ALTER TABLE `products` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `username` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` enum('active','inactive','banned') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES ('u001','alice','alice@example.com','13800000001','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u002','bob','bob@example.com','13800000002','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u003','carol','carol@example.com','13800000003','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u004','david','david@example.com','13800000004','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u005','eve','eve@example.com','13800000005','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u006','frank','frank@example.com','13800000006','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u007','grace','grace@example.com','13800000007','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u008','henry','henry@example.com','13800000008','inactive','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u009','ivy','ivy@example.com','13800000009','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u010','jack','jack@example.com','13800000010','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u011','kate','kate@example.com','13800000011','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u012','leo','leo@example.com','13800000012','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u013','mia','mia@example.com','13800000013','banned','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u014','noah','noah@example.com','13800000014','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524'),('u015','olivia','olivia@example.com','13800000015','active','2026-06-02 15:20:56.524','2026-06-02 15:20:56.524');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;


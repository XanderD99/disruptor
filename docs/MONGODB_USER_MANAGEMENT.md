# MongoDB User Management Guide 🍃

This guide explains how to create and manage MongoDB users for the Discord Disruptor project.

## 📋 **Table of Contents**

- [Default Setup](#default-setup)
- [Creating New Users](#creating-new-users)
- [User Types and Roles](#user-types-and-roles)
- [Connection Strings](#connection-strings)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)

## 🔐 **Default Setup**

After running the MongoDB Docker Compose setup, you'll have:

- **Root User**: `disruptor-admin` (full admin access)
- **Database**: `disruptor`
- **Root Password**: Generated and stored in `./secrets/mongodb_root_password.txt`

## 👤 **Creating New Users**

### **Method 1: Using MongoDB Shell (Recommended)**

#### 1. Connect to MongoDB as Root

```bash
# Get the root password
ROOT_PASSWORD=$(cat ./secrets/mongodb_root_password.txt)

# Connect to MongoDB
docker exec -it disruptor-mongodb-prod mongosh \
  --username disruptor-admin \
  --password "$ROOT_PASSWORD" \
  --authenticationDatabase admin
```

#### 2. Create Application User

```javascript
// Switch to disruptor database
use disruptor;

// Create read-write user for the bot
db.createUser({
  user: "disruptor-bot",
  pwd: "your-secure-bot-password",
  roles: [
    {
      role: "readWrite",
      db: "disruptor"
    }
  ]
});

// Verify user creation
db.getUsers();
```

#### 3. Create Read-Only User (for monitoring/analytics)

```javascript
// Create read-only user
db.createUser({
  user: "disruptor-readonly",
  pwd: "your-secure-readonly-password",
  roles: [
    {
      role: "read",
      db: "disruptor"
    }
  ]
});
```

#### 4. Create Admin User for Database Management

```javascript
// Switch to admin database
use admin;

// Create database admin user
db.createUser({
  user: "disruptor-dbadmin",
  pwd: "your-secure-admin-password",
  roles: [
    {
      role: "dbAdminAnyDatabase",
      db: "admin"
    },
    {
      role: "readWriteAnyDatabase",
      db: "admin"
    }
  ]
});
```

### **Method 2: Using Docker Exec Commands**

```bash
# Create bot user
docker exec -it disruptor-mongodb-prod mongosh \
  --username disruptor-admin \
  --password "$(cat ./secrets/mongodb_root_password.txt)" \
  --authenticationDatabase admin \
  --eval "
    use disruptor;
    db.createUser({
      user: 'disruptor-bot',
      pwd: 'your-secure-bot-password',
      roles: [{role: 'readWrite', db: 'disruptor'}]
    });
    print('Bot user created successfully');
  "

# Create readonly user
docker exec -it disruptor-mongodb-prod mongosh \
  --username disruptor-admin \
  --password "$(cat ./secrets/mongodb_root_password.txt)" \
  --authenticationDatabase admin \
  --eval "
    use disruptor;
    db.createUser({
      user: 'disruptor-readonly',
      pwd: 'your-secure-readonly-password',
      roles: [{role: 'read', db: 'disruptor'}]
    });
    print('Readonly user created successfully');
  "
```

### **Method 3: Using Initialization Scripts**

Create a user initialization script that runs on container startup:

```javascript
// filepath: ./docker/mongodb/init-scripts/02-create-app-users.js

// Switch to disruptor database
db = db.getSiblingDB('disruptor');

// Create bot user
db.createUser({
  user: 'disruptor-bot',
  pwd: 'temp-bot-password-change-me',
  roles: [
    {
      role: 'readWrite',
      db: 'disruptor'
    }
  ]
});

// Create readonly user
db.createUser({
  user: 'disruptor-readonly',
  pwd: 'temp-readonly-password-change-me',
  roles: [
    {
      role: 'read',
      db: 'disruptor'
    }
  ]
});

print('Application users created successfully');
print('IMPORTANT: Change the temporary passwords after initialization!');
```

## 🎭 **User Types and Roles**

### **Bot Application User**

- **Purpose**: Primary user for the Discord bot
- **Role**: `readWrite` on `disruptor` database
- **Permissions**: Create, read, update, delete documents and collections

### **Read-Only User**

- **Purpose**: Monitoring, analytics, backups
- **Role**: `read` on `disruptor` database
- **Permissions**: Read documents and collections only

### **Database Admin User**

- **Purpose**: Database maintenance and management
- **Role**: `dbAdminAnyDatabase`, `readWriteAnyDatabase`
- **Permissions**: Full database administration

### **Root User**

- **Purpose**: System administration
- **Role**: `root`
- **Permissions**: Full MongoDB instance control

## 🔗 **Connection Strings**

### **Bot Application User**

```bash
mongodb://disruptor-bot:your-secure-bot-password@mongodb:27017/disruptor?authSource=disruptor
```

### **Read-Only User**

```bash
mongodb://disruptor-readonly:your-secure-readonly-password@mongodb:27017/disruptor?authSource=disruptor
```

### **Root User**

```bash
mongodb://disruptor-admin:$(cat ./secrets/mongodb_root_password.txt)@mongodb:27017/disruptor?authSource=admin
```

## 🛡️ **Security Best Practices**

### **Password Management**

```bash
# Generate secure passwords
openssl rand -base64 32

# Store passwords securely
echo "your-secure-password" > ./secrets/bot_password.txt
chmod 600 ./secrets/bot_password.txt
```

### **Principle of Least Privilege**

- Use specific roles for each user type
- Don't give admin access unless necessary
- Create separate users for different applications

### **Connection Security**

```yaml
# Use Docker secrets in production
environment:
  - CONFIG_DATABASE_URL_FILE=/run/secrets/mongodb_bot_connection
secrets:
  - mongodb_bot_connection
```

## 🔧 **User Management Commands**

### **List All Users**

```javascript
// List users in current database
db.getUsers();

// List all users in all databases (admin only)
use admin;
db.runCommand({usersInfo: 1});
```

### **Update User Password**

```javascript
use disruptor;
db.updateUser("disruptor-bot", {
  pwd: "new-secure-password"
});
```

### **Grant Additional Roles**

```javascript
use disruptor;
db.grantRolesToUser("disruptor-bot", [
  {role: "dbAdmin", db: "disruptor"}
]);
```

### **Remove User**

```javascript
use disruptor;
db.dropUser("username-to-remove");
```

## 🔍 **Testing User Access**

### **Test Bot User Connection**

```bash
docker exec -it disruptor-mongodb-prod mongosh \
  --username disruptor \
  --password "your-password" \
  --authenticationDatabase disruptor \
  --eval "
    use disruptor;
    db.runCommand({ping: 1});
    print('Bot user connection successful');
  "
```

docker exec -it database-mongodb-1 mongosh --username disruptor-admin --password "yD7rGO3fZXvP6PMdczYOGD50Su5OBmHKmmpj1TPg"

### **Test Read-Only User**

```bash
docker exec -it disruptor-mongodb-prod mongosh \
  --username disruptor-readonly \
  --password "your-secure-readonly-password" \
  --authenticationDatabase disruptor \
  --eval "
    use disruptor;
    db.soundboard_configs.find().limit(1);
    print('Readonly user connection successful');
  "
```

## 🚨 **Troubleshooting**

### **Authentication Failed**

- Verify username and password
- Check authentication database (`authSource`)
- Ensure user exists: `db.getUsers()`

### **Permission Denied**

- Check user roles: `db.getUser("username")`
- Verify database permissions
- Ensure correct database context

### **Connection Issues**

```bash
# Check if MongoDB is running
docker-compose -f docker/compose.prod.yml ps mongodb

# Check MongoDB logs
docker-compose -f docker/compose.prod.yml logs mongodb

# Test basic connection
docker exec -it disruptor-mongodb-prod mongosh --eval "db.adminCommand('ping')"
```

## 📝 **Example Setup Script**

```bash
#!/bin/bash
# filepath: setup-mongodb-users.sh

echo "Setting up MongoDB users for Discord Disruptor..."

ROOT_PASSWORD=$(cat ./secrets/mongodb_root_password.txt)

# Generate secure passwords
BOT_PASSWORD=$(openssl rand -base64 32)
READONLY_PASSWORD=$(openssl rand -base64 32)

# Create bot user
docker exec -it disruptor-mongodb-prod mongosh \
  --username disruptor-admin \
  --password "$ROOT_PASSWORD" \
  --authenticationDatabase admin \
  --eval "
    use disruptor;
    db.createUser({
      user: 'disruptor-bot',
      pwd: '$BOT_PASSWORD',
      roles: [{role: 'readWrite', db: 'disruptor'}]
    });
  "

# Create readonly user
docker exec -it disruptor-mongodb-prod mongosh \
  --username disruptor-admin \
  --password "$ROOT_PASSWORD" \
  --authenticationDatabase admin \
  --eval "
    use disruptor;
    db.createUser({
      user: 'disruptor-readonly',
      pwd: '$READONLY_PASSWORD',
      roles: [{role: 'read', db: 'disruptor'}]
    });
  "

# Save passwords
echo "$BOT_PASSWORD" > ./secrets/bot_password.txt
echo "$READONLY_PASSWORD" > ./secrets/readonly_password.txt
chmod 600 ./secrets/*_password.txt

echo "✅ Users created successfully!"
echo "Bot password saved to: ./secrets/bot_password.txt"
echo "Readonly password saved to: ./secrets/readonly_password.txt"
echo ""
echo "Bot connection string:"
echo "mongodb://disruptor-bot:$BOT_PASSWORD@mongodb:27017/disruptor?authSource=disruptor"
```

Run the setup:

```bash
chmod +x setup-mongodb-users.sh
./setup-mongodb-users.sh
```

## 🎯 **Quick Reference**

| User Type | Username | Role | Use Case |
|-----------|----------|------|----------|
| Root | `disruptor-admin` | `root` | System administration |
| Bot | `disruptor-bot` | `readWrite` | Discord bot operations |
| Readonly | `disruptor-readonly` | `read` | Monitoring/analytics |
| DB Admin | `disruptor-dbadmin` | `dbAdminAnyDatabase` | Database management |

---

**🔐 Security Note**: Always use strong, unique passwords and store them securely. Never commit passwords to version control!

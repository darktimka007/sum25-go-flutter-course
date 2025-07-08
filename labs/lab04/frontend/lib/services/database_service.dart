import 'package:sqflite/sqflite.dart';
import 'package:path/path.dart';
import '../models/user.dart';

class DatabaseService {
  static Database? _database;
  static const String _dbName = 'lab04_app.db';
  static const int _version = 1;

  // Return existing database or initialize new one
  static Future<Database> get database async {
    if (_database != null) return _database!;
    _database = await _initDatabase();
    return _database!;
  }

  // Initialize the SQLite database
  static Future<Database> _initDatabase() async {
    // Get the databases path
    final databasesPath = await getDatabasesPath();
    final path = join(databasesPath, _dbName);

    // Open database with version and callbacks
    return await openDatabase(
      path,
      version: _version,
      onCreate: _onCreate,
      onUpgrade: _onUpgrade,
    );
  }

  // Create tables when database is first created
  static Future<void> _onCreate(Database db, int version) async {
    // Create users table
    await db.execute('''
      CREATE TABLE users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL
      )
    ''');

    // Create posts table with foreign key to users
    await db.execute('''
      CREATE TABLE posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        title TEXT NOT NULL,
        content TEXT,
        published INTEGER DEFAULT 0,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL,
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
      )
    ''');

    // Create index for efficient email lookups
    await db.execute('CREATE INDEX idx_users_email ON users(email)');

    // Create index for user posts lookup
    await db.execute('CREATE INDEX idx_posts_user_id ON posts(user_id)');
  }

  // Handle database schema upgrades
  static Future<void> _onUpgrade(
      Database db, int oldVersion, int newVersion) async {
    // For now, this is empty - add migration logic later if needed
  }

  // User CRUD operations

  // Insert user into database
  static Future<User> createUser(CreateUserRequest request) async {
    final db = await database;
    final now = DateTime.now().toIso8601String();

    final id = await db.insert(
      'users',
      {
        'name': request.name,
        'email': request.email,
        'created_at': now,
        'updated_at': now,
      },
      conflictAlgorithm: ConflictAlgorithm.abort,
    );

    return User(
      id: id,
      name: request.name,
      email: request.email,
      createdAt: DateTime.parse(now),
      updatedAt: DateTime.parse(now),
    );
  }

  // Get user by ID from database
  static Future<User?> getUser(int id) async {
    final db = await database;
    final List<Map<String, dynamic>> maps = await db.query(
      'users',
      where: 'id = ?',
      whereArgs: [id],
      limit: 1,
    );

    if (maps.isEmpty) {
      return null;
    }

    return User.fromMap(maps.first);
  }

  // Get all users ordered by created_at
  static Future<List<User>> getAllUsers() async {
    final db = await database;
    final List<Map<String, dynamic>> maps = await db.query(
      'users',
      orderBy: 'created_at ASC',
    );

    return List.generate(maps.length, (i) {
      return User.fromMap(maps[i]);
    });
  }

  // Update user with provided data
  static Future<User> updateUser(int id, Map<String, dynamic> updates) async {
    final db = await database;

    // Add updated_at timestamp to updates
    final updatesWithTimestamp = Map<String, dynamic>.from(updates);
    updatesWithTimestamp['updated_at'] = DateTime.now().toIso8601String();

    final count = await db.update(
      'users',
      updatesWithTimestamp,
      where: 'id = ?',
      whereArgs: [id],
    );

    if (count == 0) {
      throw Exception('User with id $id not found');
    }

    // Return updated user
    final updatedUser = await getUser(id);
    if (updatedUser == null) {
      throw Exception('Failed to retrieve updated user');
    }

    return updatedUser;
  }

  // Delete user from database
  static Future<void> deleteUser(int id) async {
    final db = await database;

    final count = await db.delete(
      'users',
      where: 'id = ?',
      whereArgs: [id],
    );

    if (count == 0) {
      throw Exception('User with id $id not found');
    }
  }

  // Count total number of users
  static Future<int> getUserCount() async {
    final db = await database;
    final result = await db.rawQuery('SELECT COUNT(*) as count FROM users');
    return result.first['count'] as int;
  }

  // Search users by name or email using LIKE operator
  static Future<List<User>> searchUsers(String query) async {
    final db = await database;
    final List<Map<String, dynamic>> maps = await db.query(
      'users',
      where: 'name LIKE ? OR email LIKE ?',
      whereArgs: ['%$query%', '%$query%'],
      orderBy: 'created_at ASC',
    );

    return List.generate(maps.length, (i) {
      return User.fromMap(maps[i]);
    });
  }

  // Database utility methods

  // Close database connection
  static Future<void> closeDatabase() async {
    if (_database != null) {
      await _database!.close();
      _database = null;
    }
  }

  // Clear all data from database (for testing)
  static Future<void> clearAllData() async {
    final db = await database;

    // Delete all records from all tables
    await db.delete('posts');
    await db.delete('users');

    // Reset auto-increment counters
    await db.delete('sqlite_sequence');
  }

  // Get the full path to the database file
  static Future<String> getDatabasePath() async {
    final databasesPath = await getDatabasesPath();
    return join(databasesPath, _dbName);
  }
}

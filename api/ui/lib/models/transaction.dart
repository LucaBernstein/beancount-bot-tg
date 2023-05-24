class Transaction {
  final int id;
  final String createdAt;
  String booking;
  bool isArchived;

  Transaction(
      {required this.id,
      required this.createdAt,
      required this.booking,
      this.isArchived = false});
}

class Config {
  final bool enableApi;
  final bool isAdmin;
  String? currency;
  String? vacationTag;
  int timezoneOffset;
  bool omitLeadingCommandSlash;

  Config(this.enableApi, this.isAdmin, this.currency, this.vacationTag,
      this.timezoneOffset, this.omitLeadingCommandSlash);
}

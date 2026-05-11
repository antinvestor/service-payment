//
//  Generated code. Do not modify.
//  source: v1/ledger.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../common/v1/common.pbjson.dart' as $8;
import '../common/v1/money.pbjson.dart' as $7;
import '../google/protobuf/struct.pbjson.dart' as $6;

@$core.Deprecated('Use ledgerTypeDescriptor instead')
const LedgerType$json = {
  '1': 'LedgerType',
  '2': [
    {'1': 'ASSET', '2': 0},
    {'1': 'LIABILITY', '2': 1},
    {'1': 'INCOME', '2': 2},
    {'1': 'EXPENSE', '2': 3},
    {'1': 'CAPITAL', '2': 4},
  ],
};

/// Descriptor for `LedgerType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List ledgerTypeDescriptor = $convert.base64Decode(
    'CgpMZWRnZXJUeXBlEgkKBUFTU0VUEAASDQoJTElBQklMSVRZEAESCgoGSU5DT01FEAISCwoHRV'
    'hQRU5TRRADEgsKB0NBUElUQUwQBA==');

@$core.Deprecated('Use transactionTypeDescriptor instead')
const TransactionType$json = {
  '1': 'TransactionType',
  '2': [
    {'1': 'NORMAL', '2': 0},
    {'1': 'REVERSAL', '2': 1},
    {'1': 'RESERVATION', '2': 2},
  ],
};

/// Descriptor for `TransactionType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List transactionTypeDescriptor = $convert.base64Decode(
    'Cg9UcmFuc2FjdGlvblR5cGUSCgoGTk9STUFMEAASDAoIUkVWRVJTQUwQARIPCgtSRVNFUlZBVE'
    'lPThAC');

@$core.Deprecated('Use transactionStatusDescriptor instead')
const TransactionStatus$json = {
  '1': 'TransactionStatus',
  '2': [
    {'1': 'PENDING', '2': 0},
    {'1': 'POSTED', '2': 1},
    {'1': 'REVERSED', '2': 2},
    {'1': 'VOIDED', '2': 3},
    {'1': 'FAILED', '2': 4},
    {'1': 'DRAFT', '2': 5},
  ],
};

/// Descriptor for `TransactionStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List transactionStatusDescriptor = $convert.base64Decode(
    'ChFUcmFuc2FjdGlvblN0YXR1cxILCgdQRU5ESU5HEAASCgoGUE9TVEVEEAESDAoIUkVWRVJTRU'
    'QQAhIKCgZWT0lERUQQAxIKCgZGQUlMRUQQBBIJCgVEUkFGVBAF');

@$core.Deprecated('Use accountTypeDescriptor instead')
const AccountType$json = {
  '1': 'AccountType',
  '2': [
    {'1': 'ACCOUNT_ASSET', '2': 0},
    {'1': 'ACCOUNT_LIABILITY', '2': 1},
    {'1': 'ACCOUNT_EQUITY', '2': 2},
    {'1': 'ACCOUNT_INCOME', '2': 3},
    {'1': 'ACCOUNT_EXPENSE', '2': 4},
    {'1': 'ACCOUNT_CONTRA_ASSET', '2': 5},
    {'1': 'ACCOUNT_CONTRA_LIABILITY', '2': 6},
    {'1': 'ACCOUNT_CONTRA_INCOME', '2': 7},
    {'1': 'ACCOUNT_CONTRA_EXPENSE', '2': 8},
    {'1': 'ACCOUNT_CLEARING', '2': 9},
    {'1': 'ACCOUNT_SUSPENSE', '2': 10},
    {'1': 'ACCOUNT_MEMO', '2': 11},
  ],
};

/// Descriptor for `AccountType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List accountTypeDescriptor = $convert.base64Decode(
    'CgtBY2NvdW50VHlwZRIRCg1BQ0NPVU5UX0FTU0VUEAASFQoRQUNDT1VOVF9MSUFCSUxJVFkQAR'
    'ISCg5BQ0NPVU5UX0VRVUlUWRACEhIKDkFDQ09VTlRfSU5DT01FEAMSEwoPQUNDT1VOVF9FWFBF'
    'TlNFEAQSGAoUQUNDT1VOVF9DT05UUkFfQVNTRVQQBRIcChhBQ0NPVU5UX0NPTlRSQV9MSUFCSU'
    'xJVFkQBhIZChVBQ0NPVU5UX0NPTlRSQV9JTkNPTUUQBxIaChZBQ0NPVU5UX0NPTlRSQV9FWFBF'
    'TlNFEAgSFAoQQUNDT1VOVF9DTEVBUklORxAJEhQKEEFDQ09VTlRfU1VTUEVOU0UQChIQCgxBQ0'
    'NPVU5UX01FTU8QCw==');

@$core.Deprecated('Use normalBalanceDescriptor instead')
const NormalBalance$json = {
  '1': 'NormalBalance',
  '2': [
    {'1': 'DEBIT', '2': 0},
    {'1': 'CREDIT', '2': 1},
    {'1': 'NONE', '2': 2},
  ],
};

/// Descriptor for `NormalBalance`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List normalBalanceDescriptor = $convert.base64Decode(
    'Cg1Ob3JtYWxCYWxhbmNlEgkKBURFQklUEAASCgoGQ1JFRElUEAESCAoETk9ORRAC');

@$core.Deprecated('Use ledgerDescriptor instead')
const Ledger$json = {
  '1': 'Ledger',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'type', '3': 2, '4': 1, '5': 14, '6': '.ledger.v1.LedgerType', '10': 'type'},
    {'1': 'parent', '3': 3, '4': 1, '5': 9, '10': 'parent'},
    {'1': 'data', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
    {'1': 'book_id', '3': 5, '4': 1, '5': 9, '10': 'bookId'},
  ],
};

/// Descriptor for `Ledger`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ledgerDescriptor = $convert.base64Decode(
    'CgZMZWRnZXISKwoCaWQYASABKAlCG7pIGHIWEAMYKDIQWzAtOWEtel8tXXszLDQwfVICaWQSKQ'
    'oEdHlwZRgCIAEoDjIVLmxlZGdlci52MS5MZWRnZXJUeXBlUgR0eXBlEhYKBnBhcmVudBgDIAEo'
    'CVIGcGFyZW50EisKBGRhdGEYBCABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0UgRkYXRhEh'
    'cKB2Jvb2tfaWQYBSABKAlSBmJvb2tJZA==');

@$core.Deprecated('Use accountDescriptor instead')
const Account$json = {
  '1': 'Account',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'ledger', '3': 3, '4': 1, '5': 9, '10': 'ledger'},
    {'1': 'balance', '3': 4, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'balance'},
    {'1': 'data', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
    {'1': 'uncleared_balance', '3': 6, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'unclearedBalance'},
    {'1': 'reserved_balance', '3': 7, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'reservedBalance'},
    {'1': 'account_type', '3': 8, '4': 1, '5': 14, '6': '.ledger.v1.AccountType', '10': 'accountType'},
    {'1': 'normal_balance', '3': 9, '4': 1, '5': 14, '6': '.ledger.v1.NormalBalance', '10': 'normalBalance'},
    {'1': 'book_id', '3': 10, '4': 1, '5': 9, '10': 'bookId'},
  ],
};

/// Descriptor for `Account`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountDescriptor = $convert.base64Decode(
    'CgdBY2NvdW50EisKAmlkGAEgASgJQhu6SBhyFhADGCgyEFswLTlhLXpfLV17Myw0MH1SAmlkEh'
    'YKBmxlZGdlchgDIAEoCVIGbGVkZ2VyEioKB2JhbGFuY2UYBCABKAsyEC5jb21tb24udjEuTW9u'
    'ZXlSB2JhbGFuY2USKwoEZGF0YRgFIAEoCzIXLmdvb2dsZS5wcm90b2J1Zi5TdHJ1Y3RSBGRhdG'
    'ESPQoRdW5jbGVhcmVkX2JhbGFuY2UYBiABKAsyEC5jb21tb24udjEuTW9uZXlSEHVuY2xlYXJl'
    'VzZXJ2ZWRCYWxhbmNlEjkKDGFjY291bnRfdHlwZRgIIAEoDjIWLmxlZGdlci52MS5BY2NvdW50'
    'VHlwZVILYWNjb3VudFR5cGUSPwoObm9ybWFsX2JhbGFuY2UYCSABKA4yGC5sZWRnZXIudjEuTm'
    '9ybWFsQmFsYW5jZVINbm9ybWFsQmFsYW5jZRIXCgdib29rX2lkGAogASgJUgZib29rSWQ=');

@$core.Deprecated('Use transactionEntryDescriptor instead')
const TransactionEntry$json = {
  '1': 'TransactionEntry',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'account_id', '3': 3, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'transaction_id', '3': 4, '4': 1, '5': 9, '10': 'transactionId'},
    {'1': 'transacted_at', '3': 5, '4': 1, '5': 9, '10': 'transactedAt'},
    {'1': 'amount', '3': 7, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'amount'},
    {'1': 'credit', '3': 8, '4': 1, '5': 8, '10': 'credit'},
    {'1': 'acc_balance', '3': 9, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'accBalance'},
    {'1': 'cleared_at', '3': 11, '4': 1, '5': 9, '10': 'clearedAt'},
  ],
};

/// Descriptor for `TransactionEntry`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transactionEntryDescriptor = $convert.base64Decode(
    'ChBUcmFuc2FjdGlvbkVudHJ5EisKAmlkGAEgASgJQhu6SBhyFhADGCgyEFswLTlhLXpfLV17My'
    'w0MH1SAmlkEh0KCmFjY291bnRfaWQYAyABKAlSCWFjY291bnRJZBIlCg50cmFuc2FjdGlvbl9p'
    'ZBgEIAEoCVINdHJhbnNhY3Rpb25JZBIjCg10cmFuc2FjdGVkX2F0GAUgASgJUgx0cmFuc2FjdG'
    'VkQXQSKAoGYW1vdW50GAcgASgLMhAuY29tbW9uLnYxLk1vbmV5UgZhbW91bnQSFgoGY3JlZGl0'
    'GAggASgIUgZjcmVkaXQSMQoLYWNjX2JhbGFuY2UYCSABKAsyEC5jb21tb24udjEuTW9uZXlSCm'
    'FjY0JhbGFuY2USHQoKY2xlYXJlZF9hdBgLIAEoCVIJY2xlYXJlZEF0');

@$core.Deprecated('Use transactionDescriptor instead')
const Transaction$json = {
  '1': 'Transaction',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'currency_code', '3': 2, '4': 1, '5': 9, '10': 'currencyCode'},
    {'1': 'transacted_at', '3': 3, '4': 1, '5': 9, '10': 'transactedAt'},
    {'1': 'data', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
    {'1': 'entries', '3': 5, '4': 3, '5': 11, '6': '.ledger.v1.TransactionEntry', '10': 'entries'},
    {'1': 'cleared', '3': 6, '4': 1, '5': 8, '10': 'cleared'},
    {'1': 'type', '3': 7, '4': 1, '5': 14, '6': '.ledger.v1.TransactionType', '10': 'type'},
    {'1': 'status', '3': 8, '4': 1, '5': 14, '6': '.ledger.v1.TransactionStatus', '10': 'status'},
    {'1': 'idempotency_key', '3': 9, '4': 1, '5': 9, '10': 'idempotencyKey'},
    {'1': 'external_ref', '3': 10, '4': 1, '5': 9, '10': 'externalRef'},
    {'1': 'source', '3': 11, '4': 1, '5': 9, '10': 'source'},
    {'1': 'book_id', '3': 12, '4': 1, '5': 9, '10': 'bookId'},
    {'1': 'posted_at', '3': 13, '4': 1, '5': 9, '10': 'postedAt'},
    {'1': 'voided_at', '3': 14, '4': 1, '5': 9, '10': 'voidedAt'},
    {'1': 'reversed_transaction_id', '3': 15, '4': 1, '5': 9, '10': 'reversedTransactionId'},
  ],
};

/// Descriptor for `Transaction`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transactionDescriptor = $convert.base64Decode(
    'CgtUcmFuc2FjdGlvbhIOCgJpZBgBIAEoCVICaWQSIwoNY3VycmVuY3lfY29kZRgCIAEoCVIMY3'
    'VycmVuY3lDb2RlEiMKDXRyYW5zYWN0ZWRfYXQYAyABKAlSDHRyYW5zYWN0ZWRBdBIrCgRkYXRh'
    'GAQgASgLMhcuZ29vZ2xlLnByb3RvYnVmLlN0cnVjdFIEZGF0YRI1CgdlbnRyaWVzGAUgAygLMh'
    'subGVkZ2VyLnYxLlRyYW5zYWN0aW9uRW50cnlSB2VudHJpZXMSGAoHY2xlYXJlZBgGIAEoCFIH'
    'Y2xlYXJlZBIuCgR0eXBlGAcgASgOMhoubGVkZ2VyLnYxLlRyYW5zYWN0aW9uVHlwZVIEdHlwZR'
    'I0CgZzdGF0dXMYCCABKA4yHC5sZWRnZXIudjEuVHJhbnNhY3Rpb25TdGF0dXNSBnN0YXR1cxIn'
    'Cg9pZGVtcG90ZW5jeV9rZXkYCSABKAlSDmlkZW1wb3RlbmN5S2V5EiEKDGV4dGVybmFsX3JlZh'
    'gKIAEoCVILZXh0ZXJuYWxSZWYSFgoGc291cmNlGAsgASgJUgZzb3VyY2USFwoHYm9va19pZBgM'
    'IAEoCVIGYm9va0lkEhsKCXBvc3RlZF9hdBgNIAEoCVIIcG9zdGVkQXQSGwoJdm9pZGVkX2F0GA'
    '4gASgJUgh2b2lkZWRBdBI2ChdyZXZlcnNlZF90cmFuc2FjdGlvbl9pZBgPIAEoCVIVcmV2ZXJz'
    'ZWRUcmFuc2FjdGlvbklk');

@$core.Deprecated('Use bookDescriptor instead')
const Book$json = {
  '1': 'Book',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'type', '3': 3, '4': 1, '5': 9, '10': 'type'},
    {'1': 'parent_id', '3': 4, '4': 1, '5': 9, '10': 'parentId'},
    {'1': 'currency', '3': 5, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'data', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `Book`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List bookDescriptor = $convert.base64Decode(
    'CgRCb29rEg4KAmlkGAEgASgJUgJpZBISCgRuYW1lGAIgASgJUgRuYW1lEhIKBHR5cGUYAyABKA'
    'lSBHR5cGUSGwoJcGFyZW50X2lkGAQgASgJUghwYXJlbnRJZBIaCghjdXJyZW5jeRgFIAEoCVII'
    'Y3VycmVuY3kSKwoEZGF0YRgGIAEoCzIXLmdvb2dsZS5wcm90b2J1Zi5TdHJ1Y3RSBGRhdGE=');

@$core.Deprecated('Use trialBalanceLineDescriptor instead')
const TrialBalanceLine$json = {
  '1': 'TrialBalanceLine',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'ledger_id', '3': 2, '4': 1, '5': 9, '10': 'ledgerId'},
    {'1': 'ledger_type', '3': 3, '4': 1, '5': 14, '6': '.ledger.v1.LedgerType', '10': 'ledgerType'},
    {'1': 'currency', '3': 4, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'total_debits', '3': 5, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'totalDebits'},
    {'1': 'total_credits', '3': 6, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'totalCredits'},
    {'1': 'net_balance', '3': 7, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'netBalance'},
  ],
};

/// Descriptor for `TrialBalanceLine`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List trialBalanceLineDescriptor = $convert.base64Decode(
    'ChBUcmlhbEJhbGFuY2VMaW5lEh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIbCglsZW'
    'RnZXJfaWQYAiABKAlSCGxlZGdlcklkEjYKC2xlZGdlcl90eXBlGAMgASgOMhUubGVkZ2VyLnYx'
    'LkxlZGdlclR5cGVSCmxlZGdlclR5cGUSGgoIY3VycmVuY3kYBCABKAlSCGN1cnJlbmN5EjMKDH'
    'RvdGFsX2RlYml0cxgFIAEoCzIQLmNvbW1vbi52MS5Nb25leVILdG90YWxEZWJpdHMSNQoNdG90'
    'YWxfY3JlZGl0cxgGIAEoCzIQLmNvbW1vbi52MS5Nb25leVIMdG90YWxDcmVkaXRzEjEKC25ldF'
    '9iYWxhbmNlGAcgASgLMhAuY29tbW9uLnYxLk1vbmV5UgpuZXRCYWxhbmNl');

@$core.Deprecated('Use trialBalanceTotalDescriptor instead')
const TrialBalanceTotal$json = {
  '1': 'TrialBalanceTotal',
  '2': [
    {'1': 'currency', '3': 1, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'total_debits', '3': 2, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'totalDebits'},
    {'1': 'total_credits', '3': 3, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'totalCredits'},
    {'1': 'is_balanced', '3': 4, '4': 1, '5': 8, '10': 'isBalanced'},
  ],
};

/// Descriptor for `TrialBalanceTotal`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List trialBalanceTotalDescriptor = $convert.base64Decode(
    'ChFUcmlhbEJhbGFuY2VUb3RhbBIaCghjdXJyZW5jeRgBIAEoCVIIY3VycmVuY3kSMwoMdG90YW'
    'xfZGViaXRzGAIgASgLMhAuY29tbW9uLnYxLk1vbmV5Ugt0b3RhbERlYml0cxI1Cg10b3RhbF9j'
    'cmVkaXRzGAMgASgLMhAuY29tbW9uLnYxLk1vbmV5Ugx0b3RhbENyZWRpdHMSHwoLaXNfYmFsYW'
    '5jZWQYBCABKAhSCmlzQmFsYW5jZWQ=');

@$core.Deprecated('Use statementEntryDescriptor instead')
const StatementEntry$json = {
  '1': 'StatementEntry',
  '2': [
    {'1': 'entry_id', '3': 1, '4': 1, '5': 9, '10': 'entryId'},
    {'1': 'transaction_id', '3': 2, '4': 1, '5': 9, '10': 'transactionId'},
    {'1': 'transacted_at', '3': 3, '4': 1, '5': 9, '10': 'transactedAt'},
    {'1': 'transaction_type', '3': 4, '4': 1, '5': 14, '6': '.ledger.v1.TransactionType', '10': 'transactionType'},
    {'1': 'amount', '3': 5, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'amount'},
    {'1': 'credit', '3': 6, '4': 1, '5': 8, '10': 'credit'},
    {'1': 'running_balance', '3': 7, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'runningBalance'},
    {'1': 'transaction_data', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'transactionData'},
    {'1': 'cleared_at', '3': 9, '4': 1, '5': 9, '10': 'clearedAt'},
  ],
};

/// Descriptor for `StatementEntry`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List statementEntryDescriptor = $convert.base64Decode(
    'Cg5TdGF0ZW1lbnRFbnRyeRIZCghlbnRyeV9pZBgBIAEoCVIHZW50cnlJZBIlCg50cmFuc2FjdG'
    'lvbl9pZBgCIAEoCVINdHJhbnNhY3Rpb25JZBIjCg10cmFuc2FjdGVkX2F0GAMgASgJUgx0cmFu'
    'c2FjdGVkQXQSRQoQdHJhbnNhY3Rpb25fdHlwZRgEIAEoDjIaLmxlZGdlci52MS5UcmFuc2FjdG'
    'lvblR5cGVSD3RyYW5zYWN0aW9uVHlwZRIoCgZhbW91bnQYBSABKAsyEC5jb21tb24udjEuTW9u'
    'ZXlSBmFtb3VudBIWCgZjcmVkaXQYBiABKAhSBmNyZWRpdBI5Cg9ydW5uaW5nX2JhbGFuY2UYBy'
    'ABKAsyEC5jb21tb24udjEuTW9uZXlSDnJ1bm5pbmdCYWxhbmNlEkIKEHRyYW5zYWN0aW9uX2Rh'
    'dGEYCCABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0Ug90cmFuc2FjdGlvbkRhdGESHQoKY2'
    'xlYXJlZF9hdBgJIAEoCVIJY2xlYXJlZEF0');

@$core.Deprecated('Use searchLedgersResponseDescriptor instead')
const SearchLedgersResponse$json = {
  '1': 'SearchLedgersResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.ledger.v1.Ledger', '10': 'data'},
  ],
};

/// Descriptor for `SearchLedgersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchLedgersResponseDescriptor = $convert.base64Decode(
    'ChVTZWFyY2hMZWRnZXJzUmVzcG9uc2USJQoEZGF0YRgBIAMoCzIRLmxlZGdlci52MS5MZWRnZX'
    'JSBGRhdGE=');

@$core.Deprecated('Use createLedgerRequestDescriptor instead')
const CreateLedgerRequest$json = {
  '1': 'CreateLedgerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'type', '3': 2, '4': 1, '5': 14, '6': '.ledger.v1.LedgerType', '10': 'type'},
    {'1': 'parent_id', '3': 3, '4': 1, '5': 9, '10': 'parentId'},
    {'1': 'data', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `CreateLedgerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createLedgerRequestDescriptor = $convert.base64Decode(
    'ChNDcmVhdGVMZWRnZXJSZXF1ZXN0Ei4KAmlkGAEgASgJQh66SBvYAQFyFhADGCgyEFswLTlhLX'
    'pfLV17Myw0MH1SAmlkEikKBHR5cGUYAiABKA4yFS5sZWRnZXIudjEuTGVkZ2VyVHlwZVIEdHlw'
    'ZRIbCglwYXJlbnRfaWQYAyABKAlSCHBhcmVudElkEisKBGRhdGEYBCABKAsyFy5nb29nbGUucH'
    'JvdG9idWYuU3RydWN0UgRkYXRh');

@$core.Deprecated('Use createLedgerResponseDescriptor instead')
const CreateLedgerResponse$json = {
  '1': 'CreateLedgerResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Ledger', '10': 'data'},
  ],
};

/// Descriptor for `CreateLedgerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createLedgerResponseDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVMZWRnZXJSZXNwb25zZRIlCgRkYXRhGAEgASgLMhEubGVkZ2VyLnYxLkxlZGdlcl'
    'IEZGF0YQ==');

@$core.Deprecated('Use updateLedgerRequestDescriptor instead')
const UpdateLedgerRequest$json = {
  '1': 'UpdateLedgerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'data', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `UpdateLedgerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateLedgerRequestDescriptor = $convert.base64Decode(
    'ChNVcGRhdGVMZWRnZXJSZXF1ZXN0EisKAmlkGAEgASgJQhu6SBhyFhADGCgyEFswLTlhLXpfLV'
    '17Myw0MH1SAmlkEisKBGRhdGEYBCABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0UgRkYXRh');

@$core.Deprecated('Use updateLedgerResponseDescriptor instead')
const UpdateLedgerResponse$json = {
  '1': 'UpdateLedgerResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Ledger', '10': 'data'},
  ],
};

/// Descriptor for `UpdateLedgerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateLedgerResponseDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVMZWRnZXJSZXNwb25zZRIlCgRkYXRhGAEgASgLMhEubGVkZ2VyLnYxLkxlZGdlcl'
    'IEZGF0YQ==');

@$core.Deprecated('Use searchAccountsResponseDescriptor instead')
const SearchAccountsResponse$json = {
  '1': 'SearchAccountsResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.ledger.v1.Account', '10': 'data'},
  ],
};

/// Descriptor for `SearchAccountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchAccountsResponseDescriptor = $convert.base64Decode(
    'ChZTZWFyY2hBY2NvdW50c1Jlc3BvbnNlEiYKBGRhdGEYASADKAsyEi5sZWRnZXIudjEuQWNjb3'
    'VudFIEZGF0YQ==');

@$core.Deprecated('Use createAccountRequestDescriptor instead')
const CreateAccountRequest$json = {
  '1': 'CreateAccountRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'ledger_id', '3': 3, '4': 1, '5': 9, '10': 'ledgerId'},
    {'1': 'currency', '3': 4, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'data', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `CreateAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAccountRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVBY2NvdW50UmVxdWVzdBIuCgJpZBgBIAEoCUIeukgb2AEBchYQAxgoMhBbMC05YS'
    '16Xy1dezMsNDB9UgJpZBIbCglsZWRnZXJfaWQYAyABKAlSCGxlZGdlcklkEhoKCGN1cnJlbmN5'
    'GAQgASgJUghjdXJyZW5jeRIrCgRkYXRhGAUgASgLMhcuZ29vZ2xlLnByb3RvYnVmLlN0cnVjdF'
    'IEZGF0YQ==');

@$core.Deprecated('Use createAccountResponseDescriptor instead')
const CreateAccountResponse$json = {
  '1': 'CreateAccountResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Account', '10': 'data'},
  ],
};

/// Descriptor for `CreateAccountResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAccountResponseDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVBY2NvdW50UmVzcG9uc2USJgoEZGF0YRgBIAEoCzISLmxlZGdlci52MS5BY2NvdW'
    '50UgRkYXRh');

@$core.Deprecated('Use updateAccountRequestDescriptor instead')
const UpdateAccountRequest$json = {
  '1': 'UpdateAccountRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'data', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `UpdateAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAccountRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVBY2NvdW50UmVxdWVzdBIrCgJpZBgBIAEoCUIbukgYchYQAxgoMhBbMC05YS16Xy'
    '1dezMsNDB9UgJpZBIrCgRkYXRhGAQgASgLMhcuZ29vZ2xlLnByb3RvYnVmLlN0cnVjdFIEZGF0'
    'YQ==');

@$core.Deprecated('Use updateAccountResponseDescriptor instead')
const UpdateAccountResponse$json = {
  '1': 'UpdateAccountResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Account', '10': 'data'},
  ],
};

/// Descriptor for `UpdateAccountResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAccountResponseDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVBY2NvdW50UmVzcG9uc2USJgoEZGF0YRgBIAEoCzISLmxlZGdlci52MS5BY2NvdW'
    '50UgRkYXRh');

@$core.Deprecated('Use searchTransactionsResponseDescriptor instead')
const SearchTransactionsResponse$json = {
  '1': 'SearchTransactionsResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.ledger.v1.Transaction', '10': 'data'},
  ],
};

/// Descriptor for `SearchTransactionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchTransactionsResponseDescriptor = $convert.base64Decode(
    'ChpTZWFyY2hUcmFuc2FjdGlvbnNSZXNwb25zZRIqCgRkYXRhGAEgAygLMhYubGVkZ2VyLnYxLl'
    'RyYW5zYWN0aW9uUgRkYXRh');

@$core.Deprecated('Use createTransactionRequestDescriptor instead')
const CreateTransactionRequest$json = {
  '1': 'CreateTransactionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'currency', '3': 2, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'transacted_at', '3': 3, '4': 1, '5': 9, '10': 'transactedAt'},
    {'1': 'data', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
    {'1': 'entries', '3': 5, '4': 3, '5': 11, '6': '.ledger.v1.TransactionEntry', '10': 'entries'},
    {'1': 'cleared', '3': 6, '4': 1, '5': 8, '10': 'cleared'},
    {'1': 'type', '3': 7, '4': 1, '5': 14, '6': '.ledger.v1.TransactionType', '10': 'type'},
  ],
};

/// Descriptor for `CreateTransactionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTransactionRequestDescriptor = $convert.base64Decode(
    'ChhDcmVhdGVUcmFuc2FjdGlvblJlcXVlc3QSLgoCaWQYASABKAlCHrpIG9gBAXIWEAMYKDIQWz'
    'AtOWEtel8tXXszLDQwfVICaWQSGgoIY3VycmVuY3kYAiABKAlSCGN1cnJlbmN5EiMKDXRyYW5z'
    'YWN0ZWRfYXQYAyABKAlSDHRyYW5zYWN0ZWRBdBIrCgRkYXRhGAQgASgLMhcuZ29vZ2xlLnByb3'
    'RvYnVmLlN0cnVjdFIEZGF0YRI1CgdlbnRyaWVzGAUgAygLMhsubGVkZ2VyLnYxLlRyYW5zYWN0'
    'aW9uRW50cnlSB2VudHJpZXMSGAoHY2xlYXJlZBgGIAEoCFIHY2xlYXJlZBIuCgR0eXBlGAcgAS'
    'gOMhoubGVkZ2VyLnYxLlRyYW5zYWN0aW9uVHlwZVIEdHlwZQ==');

@$core.Deprecated('Use createTransactionResponseDescriptor instead')
const CreateTransactionResponse$json = {
  '1': 'CreateTransactionResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Transaction', '10': 'data'},
  ],
};

/// Descriptor for `CreateTransactionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTransactionResponseDescriptor = $convert.base64Decode(
    'ChlDcmVhdGVUcmFuc2FjdGlvblJlc3BvbnNlEioKBGRhdGEYASABKAsyFi5sZWRnZXIudjEuVH'
    'JhbnNhY3Rpb25SBGRhdGE=');

@$core.Deprecated('Use reverseTransactionRequestDescriptor instead')
const ReverseTransactionRequest$json = {
  '1': 'ReverseTransactionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `ReverseTransactionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reverseTransactionRequestDescriptor = $convert.base64Decode(
    'ChlSZXZlcnNlVHJhbnNhY3Rpb25SZXF1ZXN0EisKAmlkGAEgASgJQhu6SBhyFhADGCgyEFswLT'
    'lhLXpfLV17Myw0MH1SAmlk');

@$core.Deprecated('Use reverseTransactionResponseDescriptor instead')
const ReverseTransactionResponse$json = {
  '1': 'ReverseTransactionResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Transaction', '10': 'data'},
  ],
};

/// Descriptor for `ReverseTransactionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reverseTransactionResponseDescriptor = $convert.base64Decode(
    'ChpSZXZlcnNlVHJhbnNhY3Rpb25SZXNwb25zZRIqCgRkYXRhGAEgASgLMhYubGVkZ2VyLnYxLl'
    'RyYW5zYWN0aW9uUgRkYXRh');

@$core.Deprecated('Use updateTransactionRequestDescriptor instead')
const UpdateTransactionRequest$json = {
  '1': 'UpdateTransactionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'cleared_at', '3': 2, '4': 1, '5': 9, '10': 'clearedAt'},
    {'1': 'data', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `UpdateTransactionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateTransactionRequestDescriptor = $convert.base64Decode(
    'ChhVcGRhdGVUcmFuc2FjdGlvblJlcXVlc3QSKwoCaWQYASABKAlCG7pIGHIWEAMYKDIQWzAtOW'
    'Etel8tXXszLDQwfVICaWQSHQoKY2xlYXJlZF9hdBgCIAEoCVIJY2xlYXJlZEF0EisKBGRhdGEY'
    'BCABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0UgRkYXRh');

@$core.Deprecated('Use updateTransactionResponseDescriptor instead')
const UpdateTransactionResponse$json = {
  '1': 'UpdateTransactionResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Transaction', '10': 'data'},
  ],
};

/// Descriptor for `UpdateTransactionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateTransactionResponseDescriptor = $convert.base64Decode(
    'ChlVcGRhdGVUcmFuc2FjdGlvblJlc3BvbnNlEioKBGRhdGEYASABKAsyFi5sZWRnZXIudjEuVH'
    'JhbnNhY3Rpb25SBGRhdGE=');

@$core.Deprecated('Use searchTransactionEntriesResponseDescriptor instead')
const SearchTransactionEntriesResponse$json = {
  '1': 'SearchTransactionEntriesResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.ledger.v1.TransactionEntry', '10': 'data'},
  ],
};

/// Descriptor for `SearchTransactionEntriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchTransactionEntriesResponseDescriptor = $convert.base64Decode(
    'CiBTZWFyY2hUcmFuc2FjdGlvbkVudHJpZXNSZXNwb25zZRIvCgRkYXRhGAEgAygLMhsubGVkZ2'
    'VyLnYxLlRyYW5zYWN0aW9uRW50cnlSBGRhdGE=');

@$core.Deprecated('Use createBookRequestDescriptor instead')
const CreateBookRequest$json = {
  '1': 'CreateBookRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'type', '3': 3, '4': 1, '5': 9, '10': 'type'},
    {'1': 'parent_id', '3': 4, '4': 1, '5': 9, '10': 'parentId'},
    {'1': 'currency', '3': 5, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'data', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Struct', '10': 'data'},
  ],
};

/// Descriptor for `CreateBookRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createBookRequestDescriptor = $convert.base64Decode(
    'ChFDcmVhdGVCb29rUmVxdWVzdBIuCgJpZBgBIAEoCUIeukgb2AEBchYQAxgoMhBbMC05YS16Xy'
    '1dezMsNDB9UgJpZBISCgRuYW1lGAIgASgJUgRuYW1lEhIKBHR5cGUYAyABKAlSBHR5cGUSGwoJ'
    'cGFyZW50X2lkGAQgASgJUghwYXJlbnRJZBIaCghjdXJyZW5jeRgFIAEoCVIIY3VycmVuY3kSKw'
    'oEZGF0YRgGIAEoCzIXLmdvb2dsZS5wcm90b2J1Zi5TdHJ1Y3RSBGRhdGE=');

@$core.Deprecated('Use createBookResponseDescriptor instead')
const CreateBookResponse$json = {
  '1': 'CreateBookResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Book', '10': 'data'},
  ],
};

/// Descriptor for `CreateBookResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createBookResponseDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVCb29rUmVzcG9uc2USIwoEZGF0YRgBIAEoCzIPLmxlZGdlci52MS5Cb29rUgRkYX'
    'Rh');

@$core.Deprecated('Use getBookRequestDescriptor instead')
const GetBookRequest$json = {
  '1': 'GetBookRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetBookRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBookRequestDescriptor = $convert.base64Decode(
    'Cg5HZXRCb29rUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use getBookResponseDescriptor instead')
const GetBookResponse$json = {
  '1': 'GetBookResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Book', '10': 'data'},
  ],
};

/// Descriptor for `GetBookResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBookResponseDescriptor = $convert.base64Decode(
    'Cg9HZXRCb29rUmVzcG9uc2USIwoEZGF0YRgBIAEoCzIPLmxlZGdlci52MS5Cb29rUgRkYXRh');

@$core.Deprecated('Use listBooksByTypeRequestDescriptor instead')
const ListBooksByTypeRequest$json = {
  '1': 'ListBooksByTypeRequest',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 9, '10': 'type'},
  ],
};

/// Descriptor for `ListBooksByTypeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listBooksByTypeRequestDescriptor = $convert.base64Decode(
    'ChZMaXN0Qm9va3NCeVR5cGVSZXF1ZXN0EhIKBHR5cGUYASABKAlSBHR5cGU=');

@$core.Deprecated('Use listBooksByTypeResponseDescriptor instead')
const ListBooksByTypeResponse$json = {
  '1': 'ListBooksByTypeResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.ledger.v1.Book', '10': 'data'},
  ],
};

/// Descriptor for `ListBooksByTypeResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listBooksByTypeResponseDescriptor = $convert.base64Decode(
    'ChdMaXN0Qm9va3NCeVR5cGVSZXNwb25zZRIjCgRkYXRhGAEgAygLMg8ubGVkZ2VyLnYxLkJvb2'
    'tSBGRhdGE=');

@$core.Deprecated('Use getTrialBalanceRequestDescriptor instead')
const GetTrialBalanceRequest$json = {
  '1': 'GetTrialBalanceRequest',
  '2': [
    {'1': 'currency', '3': 1, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'ledger_id', '3': 2, '4': 1, '5': 9, '10': 'ledgerId'},
    {'1': 'ledger_type', '3': 3, '4': 1, '5': 9, '10': 'ledgerType'},
    {'1': 'book_ids', '3': 4, '4': 3, '5': 9, '10': 'bookIds'},
    {'1': 'as_of', '3': 5, '4': 1, '5': 9, '10': 'asOf'},
  ],
};

/// Descriptor for `GetTrialBalanceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTrialBalanceRequestDescriptor = $convert.base64Decode(
    'ChZHZXRUcmlhbEJhbGFuY2VSZXF1ZXN0EhoKCGN1cnJlbmN5GAEgASgJUghjdXJyZW5jeRIbCg'
    'lsZWRnZXJfaWQYAiABKAlSCGxlZGdlcklkEh8KC2xlZGdlcl90eXBlGAMgASgJUgpsZWRnZXJU'
    'eXBlEhkKCGJvb2tfaWRzGAQgAygJUgdib29rSWRzEhMKBWFzX29mGAUgASgJUgRhc09m');

@$core.Deprecated('Use getTrialBalanceResponseDescriptor instead')
const GetTrialBalanceResponse$json = {
  '1': 'GetTrialBalanceResponse',
  '2': [
    {'1': 'lines', '3': 1, '4': 3, '5': 11, '6': '.ledger.v1.TrialBalanceLine', '10': 'lines'},
    {'1': 'totals', '3': 2, '4': 3, '5': 11, '6': '.ledger.v1.TrialBalanceTotal', '10': 'totals'},
  ],
};

/// Descriptor for `GetTrialBalanceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTrialBalanceResponseDescriptor = $convert.base64Decode(
    'ChdHZXRUcmlhbEJhbGFuY2VSZXNwb25zZRIxCgVsaW5lcxgBIAMoCzIbLmxlZGdlci52MS5Ucm'
    'lhbEJhbGFuY2VMaW5lUgVsaW5lcxI0CgZ0b3RhbHMYAiADKAsyHC5sZWRnZXIudjEuVHJpYWxC'
    'YWxhbmNlVG90YWxSBnRvdGFscw==');

@$core.Deprecated('Use getAccountStatementRequestDescriptor instead')
const GetAccountStatementRequest$json = {
  '1': 'GetAccountStatementRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'from', '3': 2, '4': 1, '5': 9, '10': 'from'},
    {'1': 'to', '3': 3, '4': 1, '5': 9, '10': 'to'},
    {'1': 'limit', '3': 4, '4': 1, '5': 5, '10': 'limit'},
    {'1': 'offset', '3': 5, '4': 1, '5': 5, '10': 'offset'},
  ],
};

/// Descriptor for `GetAccountStatementRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAccountStatementRequestDescriptor = $convert.base64Decode(
    'ChpHZXRBY2NvdW50U3RhdGVtZW50UmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW'
    '50SWQSEgoEZnJvbRgCIAEoCVIEZnJvbRIOCgJ0bxgDIAEoCVICdG8SFAoFbGltaXQYBCABKAVS'
    'BWxpbWl0EhYKBm9mZnNldBgFIAEoBVIGb2Zmc2V0');

@$core.Deprecated('Use getAccountStatementResponseDescriptor instead')
const GetAccountStatementResponse$json = {
  '1': 'GetAccountStatementResponse',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'currency', '3': 2, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'opening_balance', '3': 3, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'openingBalance'},
    {'1': 'closing_balance', '3': 4, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'closingBalance'},
    {'1': 'total_debits', '3': 5, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'totalDebits'},
    {'1': 'total_credits', '3': 6, '4': 1, '5': 11, '6': '.common.v1.Money', '10': 'totalCredits'},
    {'1': 'entries', '3': 7, '4': 3, '5': 11, '6': '.ledger.v1.StatementEntry', '10': 'entries'},
  ],
};

/// Descriptor for `GetAccountStatementResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAccountStatementResponseDescriptor = $convert.base64Decode(
    'ChtHZXRBY2NvdW50U3RhdGVtZW50UmVzcG9uc2USHQoKYWNjb3VudF9pZBgBIAEoCVIJYWNjb3'
    'VudElkEhoKCGN1cnJlbmN5GAIgASgJUghjdXJyZW5jeRI5Cg9vcGVuaW5nX2JhbGFuY2UYAyAB'
    'KAsyEC5jb21tb24udjEuTW9uZXlSDm9wZW5pbmdCYWxhbmNlEjkKD2Nsb3NpbmdfYmFsYW5jZR'
    'gEIAEoCzIQLmNvbW1vbi52MS5Nb25leVIOY2xvc2luZ0JhbGFuY2USMwoMdG90YWxfZGViaXRz'
    'GAUgASgLMhAuY29tbW9uLnYxLk1vbmV5Ugt0b3RhbERlYml0cxI1Cg10b3RhbF9jcmVkaXRzGA'
    'YgASgLMhAuY29tbW9uLnYxLk1vbmV5Ugx0b3RhbENyZWRpdHMSMwoHZW50cmllcxgHIAMoCzIZ'
    'LmxlZGdlci52MS5TdGF0ZW1lbnRFbnRyeVIHZW50cmllcw==');

@$core.Deprecated('Use voidTransactionRequestDescriptor instead')
const VoidTransactionRequest$json = {
  '1': 'VoidTransactionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `VoidTransactionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List voidTransactionRequestDescriptor = $convert.base64Decode(
    'ChZWb2lkVHJhbnNhY3Rpb25SZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use voidTransactionResponseDescriptor instead')
const VoidTransactionResponse$json = {
  '1': 'VoidTransactionResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Transaction', '10': 'data'},
  ],
};

/// Descriptor for `VoidTransactionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List voidTransactionResponseDescriptor = $convert.base64Decode(
    'ChdWb2lkVHJhbnNhY3Rpb25SZXNwb25zZRIqCgRkYXRhGAEgASgLMhYubGVkZ2VyLnYxLlRyYW'
    '5zYWN0aW9uUgRkYXRh');

@$core.Deprecated('Use markTransactionFailedRequestDescriptor instead')
const MarkTransactionFailedRequest$json = {
  '1': 'MarkTransactionFailedRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `MarkTransactionFailedRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List markTransactionFailedRequestDescriptor = $convert.base64Decode(
    'ChxNYXJrVHJhbnNhY3Rpb25GYWlsZWRSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use markTransactionFailedResponseDescriptor instead')
const MarkTransactionFailedResponse$json = {
  '1': 'MarkTransactionFailedResponse',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 11, '6': '.ledger.v1.Transaction', '10': 'data'},
  ],
};

/// Descriptor for `MarkTransactionFailedResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List markTransactionFailedResponseDescriptor = $convert.base64Decode(
    'Ch1NYXJrVHJhbnNhY3Rpb25GYWlsZWRSZXNwb25zZRIqCgRkYXRhGAEgASgLMhYubGVkZ2VyLn'
    'YxLlRyYW5zYWN0aW9uUgRkYXRh');

const $core.Map<$core.String, $core.dynamic> LedgerServiceBase$json = {
  '1': 'LedgerService',
  '2': [
    {
      '1': 'SearchLedgers',
      '2': '.common.v1.SearchRequest',
      '3': '.ledger.v1.SearchLedgersResponse',
      '4': {'34': 1},
      '6': true,
    },
    {'1': 'CreateLedger', '2': '.ledger.v1.CreateLedgerRequest', '3': '.ledger.v1.CreateLedgerResponse', '4': {}},
    {'1': 'UpdateLedger', '2': '.ledger.v1.UpdateLedgerRequest', '3': '.ledger.v1.UpdateLedgerResponse', '4': {}},
    {
      '1': 'SearchAccounts',
      '2': '.common.v1.SearchRequest',
      '3': '.ledger.v1.SearchAccountsResponse',
      '4': {'34': 1},
      '6': true,
    },
    {'1': 'CreateAccount', '2': '.ledger.v1.CreateAccountRequest', '3': '.ledger.v1.CreateAccountResponse', '4': {}},
    {'1': 'UpdateAccount', '2': '.ledger.v1.UpdateAccountRequest', '3': '.ledger.v1.UpdateAccountResponse', '4': {}},
    {
      '1': 'SearchTransactions',
      '2': '.common.v1.SearchRequest',
      '3': '.ledger.v1.SearchTransactionsResponse',
      '4': {'34': 1},
      '6': true,
    },
    {'1': 'CreateTransaction', '2': '.ledger.v1.CreateTransactionRequest', '3': '.ledger.v1.CreateTransactionResponse', '4': {}},
    {'1': 'ReverseTransaction', '2': '.ledger.v1.ReverseTransactionRequest', '3': '.ledger.v1.ReverseTransactionResponse', '4': {}},
    {'1': 'UpdateTransaction', '2': '.ledger.v1.UpdateTransactionRequest', '3': '.ledger.v1.UpdateTransactionResponse', '4': {}},
    {
      '1': 'SearchTransactionEntries',
      '2': '.common.v1.SearchRequest',
      '3': '.ledger.v1.SearchTransactionEntriesResponse',
      '4': {'34': 1},
      '6': true,
    },
    {'1': 'VoidTransaction', '2': '.ledger.v1.VoidTransactionRequest', '3': '.ledger.v1.VoidTransactionResponse', '4': {}},
    {'1': 'MarkTransactionFailed', '2': '.ledger.v1.MarkTransactionFailedRequest', '3': '.ledger.v1.MarkTransactionFailedResponse', '4': {}},
    {'1': 'CreateBook', '2': '.ledger.v1.CreateBookRequest', '3': '.ledger.v1.CreateBookResponse', '4': {}},
    {
      '1': 'GetBook',
      '2': '.ledger.v1.GetBookRequest',
      '3': '.ledger.v1.GetBookResponse',
      '4': {'34': 1},
    },
    {
      '1': 'ListBooksByType',
      '2': '.ledger.v1.ListBooksByTypeRequest',
      '3': '.ledger.v1.ListBooksByTypeResponse',
      '4': {'34': 1},
    },
    {
      '1': 'GetTrialBalance',
      '2': '.ledger.v1.GetTrialBalanceRequest',
      '3': '.ledger.v1.GetTrialBalanceResponse',
      '4': {'34': 1},
    },
    {
      '1': 'GetAccountStatement',
      '2': '.ledger.v1.GetAccountStatementRequest',
      '3': '.ledger.v1.GetAccountStatementResponse',
      '4': {'34': 1},
    },
  ],
  '3': {},
};

@$core.Deprecated('Use ledgerServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>> LedgerServiceBase$messageJson = {
  '.common.v1.SearchRequest': $8.SearchRequest$json,
  '.common.v1.PageCursor': $8.PageCursor$json,
  '.google.protobuf.Struct': $6.Struct$json,
  '.google.protobuf.Struct.FieldsEntry': $6.Struct_FieldsEntry$json,
  '.google.protobuf.Value': $6.Value$json,
  '.google.protobuf.ListValue': $6.ListValue$json,
  '.ledger.v1.SearchLedgersResponse': SearchLedgersResponse$json,
  '.ledger.v1.Ledger': Ledger$json,
  '.ledger.v1.CreateLedgerRequest': CreateLedgerRequest$json,
  '.ledger.v1.CreateLedgerResponse': CreateLedgerResponse$json,
  '.ledger.v1.UpdateLedgerRequest': UpdateLedgerRequest$json,
  '.ledger.v1.UpdateLedgerResponse': UpdateLedgerResponse$json,
  '.ledger.v1.SearchAccountsResponse': SearchAccountsResponse$json,
  '.ledger.v1.Account': Account$json,
  '.common.v1.Money': $7.Money$json,
  '.ledger.v1.CreateAccountRequest': CreateAccountRequest$json,
  '.ledger.v1.CreateAccountResponse': CreateAccountResponse$json,
  '.ledger.v1.UpdateAccountRequest': UpdateAccountRequest$json,
  '.ledger.v1.UpdateAccountResponse': UpdateAccountResponse$json,
  '.ledger.v1.SearchTransactionsResponse': SearchTransactionsResponse$json,
  '.ledger.v1.Transaction': Transaction$json,
  '.ledger.v1.TransactionEntry': TransactionEntry$json,
  '.ledger.v1.CreateTransactionRequest': CreateTransactionRequest$json,
  '.ledger.v1.CreateTransactionResponse': CreateTransactionResponse$json,
  '.ledger.v1.ReverseTransactionRequest': ReverseTransactionRequest$json,
  '.ledger.v1.ReverseTransactionResponse': ReverseTransactionResponse$json,
  '.ledger.v1.UpdateTransactionRequest': UpdateTransactionRequest$json,
  '.ledger.v1.UpdateTransactionResponse': UpdateTransactionResponse$json,
  '.ledger.v1.SearchTransactionEntriesResponse': SearchTransactionEntriesResponse$json,
  '.ledger.v1.VoidTransactionRequest': VoidTransactionRequest$json,
  '.ledger.v1.VoidTransactionResponse': VoidTransactionResponse$json,
  '.ledger.v1.MarkTransactionFailedRequest': MarkTransactionFailedRequest$json,
  '.ledger.v1.MarkTransactionFailedResponse': MarkTransactionFailedResponse$json,
  '.ledger.v1.CreateBookRequest': CreateBookRequest$json,
  '.ledger.v1.CreateBookResponse': CreateBookResponse$json,
  '.ledger.v1.Book': Book$json,
  '.ledger.v1.GetBookRequest': GetBookRequest$json,
  '.ledger.v1.GetBookResponse': GetBookResponse$json,
  '.ledger.v1.ListBooksByTypeRequest': ListBooksByTypeRequest$json,
  '.ledger.v1.ListBooksByTypeResponse': ListBooksByTypeResponse$json,
  '.ledger.v1.GetTrialBalanceRequest': GetTrialBalanceRequest$json,
  '.ledger.v1.GetTrialBalanceResponse': GetTrialBalanceResponse$json,
  '.ledger.v1.TrialBalanceLine': TrialBalanceLine$json,
  '.ledger.v1.TrialBalanceTotal': TrialBalanceTotal$json,
  '.ledger.v1.GetAccountStatementRequest': GetAccountStatementRequest$json,
  '.ledger.v1.GetAccountStatementResponse': GetAccountStatementResponse$json,
  '.ledger.v1.StatementEntry': StatementEntry$json,
};

/// Descriptor for `LedgerService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List ledgerServiceDescriptor = $convert.base64Decode(
    'Cg1MZWRnZXJTZXJ2aWNlErACCg1TZWFyY2hMZWRnZXJzEhguY29tbW9uLnYxLlNlYXJjaFJlcX'
    'Vlc3QaIC5sZWRnZXIudjEuU2VhcmNoTGVkZ2Vyc1Jlc3BvbnNlIuABkAIBukfIAQoHTGVkZ2Vy'
    'cxIOU2VhcmNoIGxlZGdlcnManQFTZWFyY2hlcyBmb3IgbGVkZ2VycyBpbiB0aGUgY2hhcnQgb2'
    'YgYWNjb3VudHMuIFN1cHBvcnRzIGZpbHRlcmluZyBieSBsZWRnZXIgdHlwZSwgcGFyZW50IGxl'
    'ZGdlciwgYW5kIGN1c3RvbSBwcm9wZXJ0aWVzLiBSZXR1cm5zIGEgc3RyZWFtIG9mIG1hdGNoaW'
    '5nIGxlZGdlcnMuKg1zZWFyY2hMZWRnZXJzgrUYDQoLbGVkZ2VyX3ZpZXcwARLAAgoMQ3JlYXRl'
    'TGVkZ2VyEh4ubGVkZ2VyLnYxLkNyZWF0ZUxlZGdlclJlcXVlc3QaHy5sZWRnZXIudjEuQ3JlYX'
    'RlTGVkZ2VyUmVzcG9uc2Ui7gG6R9cBCgdMZWRnZXJzEhNDcmVhdGUgYSBuZXcgbGVkZ2VyGqgB'
    'Q3JlYXRlcyBhIG5ldyBsZWRnZXIgaW4gdGhlIGNoYXJ0IG9mIGFjY291bnRzLiBMZWRnZXJzIH'
    'JlcHJlc2VudCBhY2NvdW50aW5nIGNhdGVnb3JpZXMgKEFzc2V0LCBMaWFiaWxpdHksIEluY29t'
    'ZSwgRXhwZW5zZSwgQ2FwaXRhbCkgYW5kIGNhbiBiZSBvcmdhbml6ZWQgaGllcmFyY2hpY2FsbH'
    'kuKgxjcmVhdGVMZWRnZXKCtRgPCg1sZWRnZXJfbWFuYWdlEo8CCgxVcGRhdGVMZWRnZXISHi5s'
    'ZWRnZXIudjEuVXBkYXRlTGVkZ2VyUmVxdWVzdBofLmxlZGdlci52MS5VcGRhdGVMZWRnZXJSZX'
    'Nwb25zZSK9AbpHpgEKB0xlZGdlcnMSFlVwZGF0ZSBsZWRnZXIgbWV0YWRhdGEadVVwZGF0ZXMg'
    'YW4gZXhpc3RpbmcgbGVkZ2VyJ3MgbWV0YWRhdGEgYW5kIHByb3BlcnRpZXMuIFRoZSBsZWRnZX'
    'IgdHlwZSBhbmQgcmVmZXJlbmNlIGNhbm5vdCBiZSBjaGFuZ2VkIGFmdGVyIGNyZWF0aW9uLioM'
    'dXBkYXRlTGVkZ2VygrUYDwoNbGVkZ2VyX21hbmFnZRLAAgoOU2VhcmNoQWNjb3VudHMSGC5jb2'
    '1tb24udjEuU2VhcmNoUmVxdWVzdBohLmxlZGdlci52MS5TZWFyY2hBY2NvdW50c1Jlc3BvbnNl'
    'Iu4BkAIBukfVAQoIQWNjb3VudHMSD1NlYXJjaCBhY2NvdW50cxqnAVNlYXJjaGVzIGZvciBhY2'
    'NvdW50cyBtYXRjaGluZyBzcGVjaWZpZWQgY3JpdGVyaWEuIFN1cHBvcnRzIGZpbHRlcmluZyBi'
    'eSBsZWRnZXIsIGJhbGFuY2UgcmFuZ2UsIGN1cnJlbmN5LCBhbmQgY3VzdG9tIHByb3BlcnRpZX'
    'MuIFJldHVybnMgYSBzdHJlYW0gb2YgbWF0Y2hpbmcgYWNjb3VudHMuKg5zZWFyY2hBY2NvdW50'
    'c4K1GA4KDGFjY291bnRfdmlldzABEqMCCg1DcmVhdGVBY2NvdW50Eh8ubGVkZ2VyLnYxLkNyZW'
    'F0ZUFjY291bnRSZXF1ZXN0GiAubGVkZ2VyLnYxLkNyZWF0ZUFjY291bnRSZXNwb25zZSLOAbpH'
    'tgEKCEFjY291bnRzEhRDcmVhdGUgYSBuZXcgYWNjb3VudBqEAUNyZWF0ZXMgYSBuZXcgYWNjb3'
    'VudCB3aXRoaW4gYSBsZWRnZXIuIEFjY291bnRzIHRyYWNrIGJhbGFuY2VzIChjbGVhcmVkLCB1'
    'bmNsZWFyZWQsIHJlc2VydmVkKSBhbmQgc3VwcG9ydCBtdWx0aS1jdXJyZW5jeSBvcGVyYXRpb2'
    '5zLioNY3JlYXRlQWNjb3VudIK1GBAKDmFjY291bnRfbWFuYWdlEqYCCg1VcGRhdGVBY2NvdW50'
    'Eh8ubGVkZ2VyLnYxLlVwZGF0ZUFjY291bnRSZXF1ZXN0GiAubGVkZ2VyLnYxLlVwZGF0ZUFjY2'
    '91bnRSZXNwb25zZSLRAbpHuQEKCEFjY291bnRzEhdVcGRhdGUgYWNjb3VudCBtZXRhZGF0YRqE'
    'AVVwZGF0ZXMgYW4gZXhpc3RpbmcgYWNjb3VudCdzIG1ldGFkYXRhIGFuZCBwcm9wZXJ0aWVzLi'
    'BBY2NvdW50IGJhbGFuY2VzIGFyZSB1cGRhdGVkIHRocm91Z2ggdHJhbnNhY3Rpb25zLCBub3Qg'
    'ZGlyZWN0bHkgdmlhIHRoaXMgUlBDLioNdXBkYXRlQWNjb3VudIK1GBAKDmFjY291bnRfbWFuYW'
    'dlEu0CChJTZWFyY2hUcmFuc2FjdGlvbnMSGC5jb21tb24udjEuU2VhcmNoUmVxdWVzdBolLmxl'
    'ZGdlci52MS5TZWFyY2hUcmFuc2FjdGlvbnNSZXNwb25zZSKTApACAbpH9gEKDFRyYW5zYWN0aW'
    '9ucxITU2VhcmNoIHRyYW5zYWN0aW9ucxq8AVNlYXJjaGVzIGZvciB0cmFuc2FjdGlvbnMgbWF0'
    'Y2hpbmcgc3BlY2lmaWVkIGNyaXRlcmlhLiBTdXBwb3J0cyBmaWx0ZXJpbmcgYnkgZGF0ZSByYW'
    '5nZSwgYWNjb3VudCwgY3VycmVuY3ksIGNsZWFyZWQgc3RhdHVzLCBhbmQgdHJhbnNhY3Rpb24g'
    'dHlwZS4gUmV0dXJucyBhIHN0cmVhbSBvZiBtYXRjaGluZyB0cmFuc2FjdGlvbnMuKhJzZWFyY2'
    'hUcmFuc2FjdGlvbnOCtRgSChB0cmFuc2FjdGlvbl92aWV3MAES5wIKEUNyZWF0ZVRyYW5zYWN0'
    'aW9uEiMubGVkZ2VyLnYxLkNyZWF0ZVRyYW5zYWN0aW9uUmVxdWVzdBokLmxlZGdlci52MS5Dcm'
    'VhdGVUcmFuc2FjdGlvblJlc3BvbnNlIoYCukfqAQoMVHJhbnNhY3Rpb25zEhhDcmVhdGUgYSBu'
    'ZXcgdHJhbnNhY3Rpb24arAFDcmVhdGVzIGEgbmV3IGRvdWJsZS1lbnRyeSB0cmFuc2FjdGlvbi'
    '4gVGhlIHRyYW5zYWN0aW9uIG11c3QgY29udGFpbiBhdCBsZWFzdCB0d28gZW50cmllcyB3aXRo'
    'IGJhbGFuY2VkIGRlYml0cyBhbmQgY3JlZGl0cy4gVXBkYXRlcyBhZmZlY3RlZCBhY2NvdW50IG'
    'JhbGFuY2VzIGF1dG9tYXRpY2FsbHkuKhFjcmVhdGVUcmFuc2FjdGlvboK1GBQKEnRyYW5zYWN0'
    'aW9uX21hbmFnZRLTAgoSUmV2ZXJzZVRyYW5zYWN0aW9uEiQubGVkZ2VyLnYxLlJldmVyc2VUcm'
    'Fuc2FjdGlvblJlcXVlc3QaJS5sZWRnZXIudjEuUmV2ZXJzZVRyYW5zYWN0aW9uUmVzcG9uc2Ui'
    '7wG6R9MBCgxUcmFuc2FjdGlvbnMSFVJldmVyc2UgYSB0cmFuc2FjdGlvbhqXAVJldmVyc2VzIG'
    'EgdHJhbnNhY3Rpb24gYnkgY3JlYXRpbmcgYSBuZXcgUkVWRVJTQUwgdHJhbnNhY3Rpb24gd2l0'
    'aCBpbnZlcnRlZCBlbnRyaWVzLiBUaGUgb3JpZ2luYWwgdHJhbnNhY3Rpb24gcmVtYWlucyBpbi'
    'B0aGUgbGVkZ2VyIGZvciBhdWRpdCBwdXJwb3Nlcy4qEnJldmVyc2VUcmFuc2FjdGlvboK1GBQK'
    'EnRyYW5zYWN0aW9uX21hbmFnZRLbAgoRVXBkYXRlVHJhbnNhY3Rpb24SIy5sZWRnZXIudjEuVX'
    'BkYXRlVHJhbnNhY3Rpb25SZXF1ZXN0GiQubGVkZ2VyLnYxLlVwZGF0ZVRyYW5zYWN0aW9uUmVz'
    'cG9uc2Ui+gG6R94BCgxUcmFuc2FjdGlvbnMSG1VwZGF0ZSB0cmFuc2FjdGlvbiBtZXRhZGF0YR'
    'qdAVVwZGF0ZXMgYSB0cmFuc2FjdGlvbidzIG1ldGFkYXRhIGFuZCBwcm9wZXJ0aWVzLiBUcmFu'
    'c2FjdGlvbiBlbnRyaWVzIGFuZCBhbW91bnRzIGNhbm5vdCBiZSBjaGFuZ2VkIGFmdGVyIGNyZW'
    'F0aW9uIC0gdXNlIFJldmVyc2VUcmFuc2FjdGlvbiB0byBjb3JyZWN0IGVycm9ycy4qEXVwZGF0'
    'ZVRyYW5zYWN0aW9ugrUYFAoSdHJhbnNhY3Rpb25fbWFuYWdlEpIDChhTZWFyY2hUcmFuc2FjdG'
    'lvbkVudHJpZXMSGC5jb21tb24udjEuU2VhcmNoUmVxdWVzdBorLmxlZGdlci52MS5TZWFyY2hU'
    'cmFuc2FjdGlvbkVudHJpZXNSZXNwb25zZSKsApACAbpHjwIKDFRyYW5zYWN0aW9ucxIaU2Vhcm'
    'NoIHRyYW5zYWN0aW9uIGVudHJpZXMayAFTZWFyY2hlcyBmb3IgaW5kaXZpZHVhbCB0cmFuc2Fj'
    'dGlvbiBlbnRyaWVzLiBVc2VmdWwgZm9yIGdlbmVyYXRpbmcgYWNjb3VudCBzdGF0ZW1lbnRzLC'
    'ByZWNvbmNpbGlhdGlvbiwgYW5kIGRldGFpbGVkIHRyYW5zYWN0aW9uIGFuYWx5c2lzLiBTdXBw'
    'b3J0cyBmaWx0ZXJpbmcgYnkgYWNjb3VudCwgZGF0ZSByYW5nZSwgYW5kIGNsZWFyZWQgc3RhdH'
    'VzLioYc2VhcmNoVHJhbnNhY3Rpb25FbnRyaWVzgrUYEgoQdHJhbnNhY3Rpb25fdmlldzABEvIC'
    'Cg9Wb2lkVHJhbnNhY3Rpb24SIS5sZWRnZXIudjEuVm9pZFRyYW5zYWN0aW9uUmVxdWVzdBoiLm'
    'xlZGdlci52MS5Wb2lkVHJhbnNhY3Rpb25SZXNwb25zZSKXArpH+wEKDFRyYW5zYWN0aW9ucxIh'
    'Vm9pZCBhIG5vdC15ZXQtcG9zdGVkIHRyYW5zYWN0aW9uGrYBVHJhbnNpdGlvbnMgYSBkcmFmdC'
    'BvciBwZW5kaW5nIHRyYW5zYWN0aW9uIHRvIFZPSURFRCBhbmQgc3RhbXBzIHZvaWRlZF9hdC4g'
    'UG9zdGVkIHRyYW5zYWN0aW9ucyBjYW5ub3QgYmUgdm9pZGVkOyB1c2UgUmV2ZXJzZVRyYW5zYW'
    'N0aW9uIHNvIHRoZSBib29rcyBjYXJyeSB0aGUgb2Zmc2V0IGFuZCBhdWRpdCB0cmFpbC4qD3Zv'
    'aWRUcmFuc2FjdGlvboK1GBQKEnRyYW5zYWN0aW9uX21hbmFnZRKOAwoVTWFya1RyYW5zYWN0aW'
    '9uRmFpbGVkEicubGVkZ2VyLnYxLk1hcmtUcmFuc2FjdGlvbkZhaWxlZFJlcXVlc3QaKC5sZWRn'
    'ZXIudjEuTWFya1RyYW5zYWN0aW9uRmFpbGVkUmVzcG9uc2UioQK6R4UCCgxUcmFuc2FjdGlvbn'
    'MSIU1hcmsgYSBwZW5kaW5nIHRyYW5zYWN0aW9uIGZhaWxlZBq6AVRyYW5zaXRpb25zIGEgcGVu'
    'ZGluZyB0cmFuc2FjdGlvbiB0byBGQUlMRUQuIFVzZSB3aGVuIHRoZSB1cHN0cmVhbSBwYXltZW'
    '50IHByb3ZpZGVyIHJlamVjdHMgdGhlIHBvc3RpbmcuIFBvc3RlZCwgdm9pZGVkLCByZXZlcnNl'
    'ZCBhbmQgYWxyZWFkeS1mYWlsZWQgdHJhbnNhY3Rpb25zIGNhbm5vdCBiZSBtYXJrZWQgZmFpbG'
    'VkLioVbWFya1RyYW5zYWN0aW9uRmFpbGVkgrUYFAoSdHJhbnNhY3Rpb25fbWFuYWdlEocDCgpD'
    'cmVhdGVCb29rEhwubGVkZ2VyLnYxLkNyZWF0ZUJvb2tSZXF1ZXN0Gh0ubGVkZ2VyLnYxLkNyZW'
    'F0ZUJvb2tSZXNwb25zZSK7ArpHpgIKBUJvb2tzEhFDcmVhdGUgYSBuZXcgYm9vaxr9AUNyZWF0'
    'ZXMgYSBuZXcgYWNjb3VudGluZyBzY29wZS4gQm9va3MgcmVwcmVzZW50IGluZGVwZW5kZW50IG'
    'FjY291bnRpbmcgZW50aXRpZXMgKHBsYXRmb3JtIGJvb2ssIGdyb3VwIGJvb2ssIG1lcmNoYW50'
    'IGJvb2ssIGV0Yy4pLiBPcHRpb25hbCBwYXJlbnRfaWQgc3VwcG9ydHMgaGllcmFyY2h5IGZvci'
    'Bjb25zb2xpZGF0ZWQgcmVwb3J0aW5nOyBjcm9zcy1ib29rIHBvc3RpbmdzIGFyZSByZWplY3Rl'
    'ZCByZWdhcmRsZXNzIG9mIGhpZXJhcmNoeS4qCmNyZWF0ZUJvb2uCtRgNCgtib29rX21hbmFnZR'
    'K8AQoHR2V0Qm9vaxIZLmxlZGdlci52MS5HZXRCb29rUmVxdWVzdBoaLmxlZGdlci52MS5HZXRC'
    'b29rUmVzcG9uc2UiepACAbpHZQoFQm9va3MSEEdldCBhIGJvb2sgYnkgaWQaQVJldHJpZXZlcy'
    'B0aGUgYm9vayB3aXRoIHRoZSBnaXZlbiBpZCB3aXRoaW4gdGhlIGNhbGxlcidzIHRlbmFuY3ku'
    'KgdnZXRCb29rgrUYCwoJYm9va192aWV3EqICCg9MaXN0Qm9va3NCeVR5cGUSIS5sZWRnZXIudj'
    'EuTGlzdEJvb2tzQnlUeXBlUmVxdWVzdBoiLmxlZGdlci52MS5MaXN0Qm9va3NCeVR5cGVSZXNw'
    'b25zZSLHAZACAbpHsQEKBUJvb2tzEhJMaXN0IGJvb2tzIGJ5IHR5cGUaggFSZXR1cm5zIGFsbC'
    'BhY3RpdmUgYm9va3Mgb2YgdGhlIHN1cHBsaWVkIGNvbnZlbnRpb25hbCB0eXBlIChlLmcuIGFs'
    'bCBncm91cCBib29rcywgYWxsIG1lcmNoYW50IGJvb2tzKSB3aXRoaW4gdGhlIGNhbGxlcidzIH'
    'RlbmFuY3kuKg9saXN0Qm9va3NCeVR5cGWCtRgLCglib29rX3ZpZXcS9wIKD0dldFRyaWFsQmFs'
    'YW5jZRIhLmxlZGdlci52MS5HZXRUcmlhbEJhbGFuY2VSZXF1ZXN0GiIubGVkZ2VyLnYxLkdldF'
    'RyaWFsQmFsYW5jZVJlc3BvbnNlIpwCkAIBukeEAgoHUmVwb3J0cxIRR2V0IHRyaWFsIGJhbGFu'
    'Y2Ua1AFDb21wdXRlcyB0aGUgdHJpYWwgYmFsYW5jZSBhY3Jvc3MgY2xlYXJlZCBOT1JNQUwgYW'
    '5kIFJFVkVSU0FMIHRyYW5zYWN0aW9ucy4gT3B0aW9uYWwgZmlsdGVycyBzY29wZSB0byBjdXJy'
    'ZW5jeSwgbGVkZ2VyLCBsZWRnZXIgdHlwZSwgYm9vayBzZXQgKGV4cGFuZCB2aWEgTGlzdEJvb2'
    'tEZXNjZW5kYW50cyBmb3IgY29uc29saWRhdGlvbiksIGFuZCBhbiBhcy1vZiBkYXRlLioPZ2V0'
    'VHJpYWxCYWxhbmNlgrUYDQoLcmVwb3J0X3ZpZXcS1gIKE0dldEFjY291bnRTdGF0ZW1lbnQSJS'
    '5sZWRnZXIudjEuR2V0QWNjb3VudFN0YXRlbWVudFJlcXVlc3QaJi5sZWRnZXIudjEuR2V0QWNj'
    'b3VudFN0YXRlbWVudFJlc3BvbnNlIu8BkAIBukfXAQoHUmVwb3J0cxIYR2V0IGFuIGFjY291bn'
    'Qgc3RhdGVtZW50GpwBUmV0dXJucyB0aGUgYWNjb3VudCdzIGVudHJpZXMgaW4gY2hyb25vbG9n'
    'aWNhbCBvcmRlciB3aXRoIHJ1bm5pbmcgYmFsYW5jZSBwZXIgZW50cnksIG9wZW5pbmcgYW5kIG'
    'Nsb3NpbmcgYmFsYW5jZXMsIGFuZCB0aGUgcGVyaW9kJ3MgcmF3IGRlYml0L2NyZWRpdCB0b3Rh'
    'bHMuKhNnZXRBY2NvdW50U3RhdGVtZW50grUYDQoLcmVwb3J0X3ZpZXcapwaCtRiiBgoOc2Vydm'
    'ljZV9sZWRnZXISC2xlZGdlcl92aWV3Eg1sZWRnZXJfbWFuYWdlEgxhY2NvdW50X3ZpZXcSDmFj'
    'Y291bnRfbWFuYWdlEhB0cmFuc2FjdGlvbl92aWV3EhJ0cmFuc2FjdGlvbl9tYW5hZ2USCWJvb2'
    'tfdmlldxILYm9va19tYW5hZ2USC3JlcG9ydF92aWV3GocBCAESC2xlZGdlcl92aWV3Eg1sZWRn'
    'ZXJfbWFuYWdlEgxhY2NvdW50X3ZpZXcSDmFjY291bnRfbWFuYWdlEhB0cmFuc2FjdGlvbl92aW'
    'V3EhJ0cmFuc2FjdGlvbl9tYW5hZ2USCWJvb2tfdmlldxILYm9va19tYW5hZ2USC3JlcG9ydF92'
    'aWV3GocBCAISC2xlZGdlcl92aWV3Eg1sZWRnZXJfbWFuYWdlEgxhY2NvdW50X3ZpZXcSDmFjY2'
    '91bnRfbWFuYWdlEhB0cmFuc2FjdGlvbl92aWV3EhJ0cmFuc2FjdGlvbl9tYW5hZ2USCWJvb2tf'
    'dmlldxILYm9va19tYW5hZ2USC3JlcG9ydF92aWV3GlsIAxILbGVkZ2VyX3ZpZXcSDGFjY291bn'
    'RfdmlldxIQdHJhbnNhY3Rpb25fdmlldxISdHJhbnNhY3Rpb25fbWFuYWdlEglib29rX3ZpZXcS'
    'C3JlcG9ydF92aWV3GkcIBBILbGVkZ2VyX3ZpZXcSDGFjY291bnRfdmlldxIQdHJhbnNhY3Rpb2'
    '5fdmlldxIJYm9va192aWV3EgtyZXBvcnRfdmlldxpHCAUSC2xlZGdlcl92aWV3EgxhY2NvdW50'
    'X3ZpZXcSEHRyYW5zYWN0aW9uX3ZpZXcSCWJvb2tfdmlldxILcmVwb3J0X3ZpZXcahwEIBhILbG'
    'VkZ2VyX3ZpZXcSDWxlZGdlcl9tYW5hZ2USDGFjY291bnRfdmlldxIOYWNjb3VudF9tYW5hZ2US'
    'EHRyYW5zYWN0aW9uX3ZpZXcSEnRyYW5zYWN0aW9uX21hbmFnZRIJYm9va192aWV3Egtib29rX2'
    '1hbmFnZRILcmVwb3J0X3ZpZXc=');


// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { Namespace, Context } from "@ory/keto-namespace-types"

class profile_user implements Namespace {}

class tenancy_access implements Namespace {
  related: {
    member: (profile_user | tenancy_access)[]
    service: profile_user[]
  }
}

class service_ledger implements Namespace {
  related: {
    owner: profile_user[]
    admin: profile_user[]
    operator: profile_user[]
    viewer: profile_user[]
    member: profile_user[]
    service: (profile_user | tenancy_access)[]

    granted_ledger_view: (profile_user | service_ledger)[]
    granted_ledger_manage: (profile_user | service_ledger)[]
    granted_account_view: (profile_user | service_ledger)[]
    granted_account_manage: (profile_user | service_ledger)[]
    granted_transaction_view: (profile_user | service_ledger)[]
    granted_transaction_manage: (profile_user | service_ledger)[]
  }

  permits = {
    ledger_view: (ctx: Context): boolean =>
      this.related.admin.includes(ctx.subject) ||
      this.related.member.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.service.includes(ctx.subject) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.granted_ledger_view.includes(ctx.subject),

    ledger_manage: (ctx: Context): boolean =>
      this.related.admin.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.service.includes(ctx.subject) ||
      this.related.granted_ledger_manage.includes(ctx.subject),

    account_view: (ctx: Context): boolean =>
      this.related.admin.includes(ctx.subject) ||
      this.related.member.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.service.includes(ctx.subject) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.granted_account_view.includes(ctx.subject),

    account_manage: (ctx: Context): boolean =>
      this.related.admin.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.service.includes(ctx.subject) ||
      this.related.granted_account_manage.includes(ctx.subject),

    transaction_view: (ctx: Context): boolean =>
      this.related.admin.includes(ctx.subject) ||
      this.related.member.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.service.includes(ctx.subject) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.granted_transaction_view.includes(ctx.subject),

    transaction_manage: (ctx: Context): boolean =>
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.service.includes(ctx.subject) ||
      this.related.granted_transaction_manage.includes(ctx.subject),
  }
}

// @flow
/* eslint-disable */
import {type PayloadType, type ReturnType} from '../constants/types/more'
import * as Constants from '../constants/gregor'
import * as RPCTypes from '../constants/types/flow-types'
import * as GregorTypes from '../constants/types/flow-types-gregor'

// Constants
export const pushState = 'gregor:pushState'

// Action Creators
export const createPushState = (payload: {|state: GregorTypes.State, reason: RPCTypes.PushReason|}) => ({
  type: pushState,
  error: false,
  payload,
})

// Action Payloads
export type PushStatePayload = ReturnType<typeof createPushState>

// All Actions
// prettier-ignore
export type Actions =
  | ReturnType<typeof createPushState>

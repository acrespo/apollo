package io.muun.apollo.domain.selector

import io.muun.apollo.data.preferences.AuthRepository
import io.muun.apollo.data.preferences.UserRepository
import io.muun.common.model.SessionStatus
import javax.inject.Inject


class LoginAuthorizedSelector @Inject constructor(
    val authRepository: AuthRepository,
    val userRepository: UserRepository
) {

    fun watch() =
        authRepository.watchSessionStatus()
            .map { it.orElse(null) }
            .map { it == SessionStatus.AUTHORIZED_BY_EMAIL }
}
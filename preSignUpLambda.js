import { inspect } from 'util'; 

export const handler = async (event, context) => {
    console.log(`Received Pre Sign-up event: ${inspect(event, { depth: null })}`);

    event.response.autoConfirmUser = false;

    event.response.autoVerifyEmail = false;

    // --- Optional: Custom Logic Examples ---
    const userEmail = event.request.userAttributes?.email;
    if (userEmail) {
        console.log(`New user '${event.userName}' (email: ${userEmail}) requires manual confirmation.`);
    } else {
        console.log(`New user '${event.userName}' requires manual confirmation (no email provided).`);
    }

    console.log(`Returning event with response: ${inspect(event.response, { depth: null })}`);

    return event;
};